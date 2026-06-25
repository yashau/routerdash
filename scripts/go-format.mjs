import { spawnSync } from "node:child_process";
import { readdir } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const ignored = new Set(["node_modules", ".svelte-kit", "build", "dist", ".git", "screenshots"]);
const files = [];

async function walk(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (ignored.has(entry.name)) continue;
    const file = join(dir, entry.name);
    if (entry.isDirectory()) {
      await walk(file);
    } else if (entry.name.endsWith(".go")) {
      files.push(file);
    }
  }
}

await walk(root);
if (files.length === 0) process.exit(0);

const result = spawnSync("gofmt", ["-w", ...files], { cwd: root, stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status ?? 0);
