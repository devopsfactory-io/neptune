// Renders neptune-readme.html to neptune-readme.gif.
// Usage: NODE_PATH=<dir with playwright-core in node_modules> node docs/assets/readme/render-gif.cjs [--frames <dir>]
// Needs ffmpeg on PATH and Chrome/Chromium (CHROME_PATH overrides the default macOS path).
const { chromium } = require('playwright-core');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const dir = __dirname;
const FPS = 12, WIDTH = 960;
const CHECK_AT = [1.6, 4.5, 7.5, 11, 15, 18.5];
const framesDir = process.argv.includes('--frames') ? process.argv[process.argv.indexOf('--frames') + 1] : null;
const chrome = process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';

(async () => {
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'neptune-gif-'));
  const browser = await chromium.launch({ executablePath: chrome, headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  await page.goto('file://' + path.join(dir, 'neptune-readme.html'));
  await page.waitForFunction(() => [...document.images].every(i => i.complete && i.naturalWidth > 0));
  const duration = await page.evaluate(() => window.DURATION);
  const shot = async (t, file) => { await page.evaluate(x => window.seek(x), t); await page.screenshot({ path: file }); };
  if (framesDir) for (const t of CHECK_AT) await shot(t, path.join(framesDir, `gif-${String(t).replace('.', '_')}.png`));
  const n = Math.round(duration * FPS);
  for (let i = 0; i < n; i++) await shot(i / FPS, path.join(tmp, `f${String(i).padStart(4, '0')}.png`));
  await browser.close();

  // Two-pass palette keeps the GIF small and the text sharp; loops forever.
  const out = path.join(dir, 'neptune-readme.gif');
  execFileSync('ffmpeg', ['-y', '-v', 'error', '-framerate', String(FPS), '-i', path.join(tmp, 'f%04d.png'),
    '-vf', `scale=${WIDTH}:-1:flags=lanczos,split[a][b];[a]palettegen=max_colors=128:stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=4:diff_mode=rectangle`,
    '-loop', '0', out], { stdio: 'inherit' });
  fs.rmSync(tmp, { recursive: true });
  console.log('wrote', out, `${n} frames`);
})().catch(e => { console.error(e); process.exit(1); });
