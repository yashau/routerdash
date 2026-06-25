import { spawnSync } from "node:child_process";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const mode = process.argv.includes("--check") ? "--check" : "--write";
const oxfmt = join(
  root,
  "web",
  "node_modules",
  ".bin",
  process.platform === "win32" ? "oxfmt.cmd" : "oxfmt",
);
const args = [
  mode,
  "--config",
  join(root, "web", ".oxfmtrc.json"),
  ".mise.toml",
  "scripts/**/*.mjs",
];
const executable = process.platform === "win32" ? "cmd.exe" : oxfmt;
const commandArgs = process.platform === "win32" ? ["/d", "/s", "/c", oxfmt, ...args] : args;
const result = spawnSync(executable, commandArgs, { cwd: root, stdio: "inherit" });

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 0);
