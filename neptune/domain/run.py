from typing import List
from dataclasses import dataclass

@dataclass
class RunOutput:
    """Output of the run steps
    - command: The command that was run
    - output: The output of the command
    - error: The error of the command
    - status: The status of the command

    Additional methods:
    - to_string: Convert the run output to a string, usually for logging
    """
    command: str
    output: str
    error: str
    status: int

    def to_string(self) -> str:
        """Convert the run output to a string"""
        newline = chr(10)
        stdout = self.output[:100].replace(newline, " ")
        stderr = self.error[:100].replace(newline, " ")
        return f"""
  [bold]- Command:[/bold] {self.command}
    [bold]- Status:[/bold] {self.status}
    [bold]- Stdout:[/bold] {stdout}...
    [bold]- Stderr:[/bold] {stderr}...
    """

@dataclass
class StepsOutput:
    """The relation of RunOutputs for each step
    - phase: The phase of the steps
    - overall_status: The overall status of the steps
    - outputs: The outputs of the steps(RunOutput)
    """
    phase: str
    overall_status: int
    outputs: List[RunOutput]
