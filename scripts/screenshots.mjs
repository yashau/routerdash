import { spawn, spawnSync } from "node:child_process";
import { createRequire } from "node:module";
import { createServer } from "node:net";
import { mkdir, rm } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const require = createRequire(join(root, "web", "package.json"));
const screenshotDir = resolve(root, process.env.ROUTERDASH_SCREENSHOTS_DIR ?? "screenshots");
const colorScheme = process.env.ROUTERDASH_SCREENSHOTS_COLOR_SCHEME ?? "light";
const viewport = {
  width: Number(process.env.ROUTERDASH_SCREENSHOTS_WIDTH ?? 1120),
  height: Number(process.env.ROUTERDASH_SCREENSHOTS_HEIGHT ?? 900),
};

const pages = [
  { path: "/", name: "dashboard" },
  { path: "/tailscale", name: "tailscale" },
  { path: "/firewall", name: "firewall" },
  { path: "/routes", name: "routes" },
  { path: "/frr", name: "frr" },
  { path: "/diagnostics", name: "diagnostics", setup: runDiagnostic },
];

function run(command, args) {
  const executable = process.platform === "win32" && command === "pnpm" ? "cmd.exe" : command;
  const commandArgs =
    process.platform === "win32" && command === "pnpm"
      ? ["/d", "/s", "/c", command, ...args]
      : args;
  const result = spawnSync(executable, commandArgs, { cwd: root, stdio: "inherit" });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}

function localBinary() {
  return join(root, "dist", process.platform === "win32" ? "routerdash.exe" : "routerdash");
}

async function findPort() {
  return await new Promise((resolvePort, reject) => {
    const server = createServer();
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      server.close(() => {
        if (address && typeof address === "object") resolvePort(address.port);
        else reject(new Error("Could not allocate a local port"));
      });
    });
  });
}

async function waitForHealth(baseUrl) {
  const deadline = Date.now() + 10000;
  while (Date.now() < deadline) {
    try {
      const response = await fetch(`${baseUrl}/healthz`);
      if (response.ok) return;
    } catch {
      await new Promise((resolveWait) => setTimeout(resolveWait, 150));
    }
  }
  throw new Error(`Timed out waiting for ${baseUrl}/healthz`);
}

async function launchChromium(chromium) {
  try {
    return await chromium.launch();
  } catch (error) {
    if (!String(error).includes("Executable doesn't exist")) throw error;
    run("pnpm", ["--dir", "web", "exec", "playwright", "install", "chromium"]);
    return await chromium.launch();
  }
}

async function runDiagnostic(page) {
  await page.getByRole("button", { name: "Run" }).click();
  await page.waitForFunction(() => document.body.textContent?.includes("icmp_seq="), {
    timeout: 5000,
  });
}

async function contentClip(page) {
  return await page.evaluate(() => {
    const elements = Array.from(document.querySelectorAll("header, main, main > *"));
    const bottom = Math.max(
      ...elements.map((element) => element.getBoundingClientRect().bottom),
      0,
    );
    const width = document.documentElement.clientWidth;
    const height = Math.ceil(
      Math.min(Math.max(bottom + 16, 160), document.documentElement.scrollHeight),
    );
    return { x: 0, y: 0, width, height };
  });
}

async function main() {
  await rm(screenshotDir, { recursive: true, force: true });
  await mkdir(screenshotDir, { recursive: true });

  const port = await findPort();
  const baseUrl = `http://127.0.0.1:${port}`;
  const server = spawn(localBinary(), {
    cwd: root,
    env: { ...process.env, ROUTERDASH_FAKE: "1", ROUTERDASH_ADDR: `127.0.0.1:${port}` },
    stdio: "ignore",
    windowsHide: true,
  });

  try {
    await waitForHealth(baseUrl);
    const { chromium } = require("playwright");
    const browser = await launchChromium(chromium);
    try {
      const context = await browser.newContext({ viewport, colorScheme, deviceScaleFactor: 1 });
      const page = await context.newPage();
      for (const item of pages) {
        await page.goto(`${baseUrl}${item.path}`, { waitUntil: "networkidle" });
        await page.waitForFunction(() => document.body.textContent?.includes("routerdash-lab"), {
          timeout: 5000,
        });
        if (item.setup) await item.setup(page);
        await page.waitForTimeout(250);
        await page.screenshot({
          path: join(screenshotDir, `${item.name}.png`),
          clip: await contentClip(page),
        });
      }
      await context.close();
    } finally {
      await browser.close();
    }
  } finally {
    server.kill();
  }

  console.log(`Screenshots written to ${screenshotDir}`);
}

await main();
