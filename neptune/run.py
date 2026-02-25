import logging
import io
import selectors
import sys
from os import getenv, environ
from subprocess import Popen, PIPE
from neptune.domain.config import NeptuneConfig
from neptune.domain.run import RunOutput, StepsOutput
from neptune.lock import LockFileInterface, WorkflowStatus
from rich.console import Console # pylint: disable=import-error
from rich.panel import Panel # pylint: disable=import-error
from rich.progress import Progress, SpinnerColumn, TextColumn # pylint: disable=import-error

LOG_ERROR_PREFIX = "neptune run error -"
LOG_PREFIX = "neptune run -"

logger = logging.getLogger(__name__)
text_formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler = logging.StreamHandler()
handler.setFormatter(text_formatter)
logger.addHandler(handler)
loglevel = getattr(logging, getenv("NEPTUNE_LOG_LEVEL", "INFO").upper(), None)
logger.setLevel(loglevel)

console = Console()

class RunSteps:
    """Run steps"""
    def __init__(self, config: NeptuneConfig, phase: str, locks: LockFileInterface):
        self.config = config
        self.lock_interface = locks
        self.phase = phase
        self.stacks = self.lock_interface.terraform_stacks.stacks
        self.workflow_phase = self.config.workflows.workflows[self.config.repository.allowed_workflow].phases[self.phase]
        self.steps = self.workflow_phase.steps

    def _run_command(self, command: str) -> RunOutput:
        """Run a command and return its output"""
        logger.info("%s Running command: %s", LOG_PREFIX, command)
        console.print(Panel.fit(
            f"[bold]Neptune is running the following command:[/bold] {command}",
            title="🌊 Neptune Runner",
            border_style="blue"
        ))
        
        buf_stdout = io.StringIO()
        def handle_stdout(stream, mask):
            """Handle output from a stream"""
            line = stream.readline()
            buf_stdout.write(line)
            sys.stdout.write(line)
        
        buf_stderr = io.StringIO()
        def handle_stderr(stream, mask):
            """Handle output from a stream"""
            line = stream.readline()
            buf_stderr.write(line)
            sys.stderr.write(line)
        
        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            task = progress.add_task(f"[white]Running: {command}", total=None)
            proc = Popen(
                command,
                shell=True,
                stdin=PIPE,
                stdout=PIPE,
                stderr=PIPE,
                text=True,
                bufsize=1,
                universal_newlines=True,
            )

            selector = selectors.DefaultSelector()
            selector.register(proc.stdout, selectors.EVENT_READ, handle_stdout)
            selector.register(proc.stderr, selectors.EVENT_READ, handle_stderr)

            while proc.poll() is None:
                events = selector.select()
                for key, mask in events:
                    callback = key.data
                    callback(key.fileobj, mask)
            
            return_code = proc.wait()
            selector.close()
        
            # Update progress
            progress.update(task, completed=True)

        # Join all captured lines
        stdout = buf_stdout.getvalue()
        stderr = buf_stderr.getvalue()
        buf_stdout.close()
        buf_stderr.close()

        run_output = RunOutput(command=command, output=stdout, error=stderr, status=return_code)

        if return_code != 0:
            logger.error("%s Command failed with return code %d", LOG_ERROR_PREFIX, return_code)
            # logger.error("%s stdout output: %s", LOG_ERROR_PREFIX, stdout)
            # logger.error("%s stderr output: %s", LOG_ERROR_PREFIX, stderr)
            return run_output
        logger.info("%s Command completed with return code %d", LOG_PREFIX, return_code)
        # logger.info("%s stdout output: %s", LOG_PREFIX, stdout)
        # logger.info("%s stderr output: %s", LOG_PREFIX, stderr)
        return run_output


    def execute(self) -> StepsOutput:
        """Execute all steps in the workflow phase"""
        logger.info("%s Executing workflow phase: %s", LOG_PREFIX, self.phase)
        
        # Update status to in progress
        self.lock_interface.update_stacks(
            phase=self.phase,
            stacks=self.stacks,
            status=WorkflowStatus.IN_PROGRESS
        )

        steps_output = StepsOutput(phase=self.phase, overall_status=0, outputs=[])
        for step in self.steps:
            result = self._run_command(step.run)

            steps_output.outputs.append(result)
            if result.status != 0:
                # If any step fails, the overall status is 1
                steps_output.overall_status = 1
                self.lock_interface.update_stacks(
                    phase=self.phase,
                    stacks=self.stacks,
                    status=WorkflowStatus.PENDING
                )
                break

        # if all items in steps_output.outputs have status 0, update the status to completed
        if all(output.status == 0 for output in steps_output.outputs):
            self.lock_interface.update_stacks(
                phase=self.phase,
                stacks=self.stacks,
                status=WorkflowStatus.COMPLETED
            )

        return steps_output
