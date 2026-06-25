import { readdir, readFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const max = 1000;
const ignored = new Set(["node_modules", ".svelte-kit", "build", "dist", ".git"]);
const extensions = new Set([".go", ".svelte", ".ts", ".js", ".mjs", ".json", ".css", ".toml"]);
const offenders = [];

async function walk(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (ignored.has(entry.name)) continue;
    const file = join(dir, entry.name);
    if (entry.isDirectory()) {
      await walk(file);
      continue;
    }
    if (!extensions.has(file.slice(file.lastIndexOf(".")))) continue;
    const text = await readFile(file, "utf8");
    const lines = text.split("\n").length;
    if (lines > max) offenders.push(`${relative(root, file)}: ${lines}`);
  }
}

await walk(root);
if (offenders.length) {
  console.error(`Files over ${max} lines:\n${offenders.join("\n")}`);
  process.exit(1);
}
