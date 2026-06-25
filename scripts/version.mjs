import { readFile, writeFile } from "node:fs/promises";

const command = process.argv[2] ?? "show";
const versionPath = "VERSION";
const pattern = /^\d{4}\.\d{2}\.\d{2}-\d+$/;

if (command === "show") {
  console.log(await readVersion());
} else if (command === "check") {
  await readVersion();
} else if (command === "bump") {
  console.log(await bumpVersion());
} else {
  console.error(`Unknown version command: ${command}`);
  process.exit(1);
}

async function readVersion() {
  const version = (await readFile(versionPath, "utf8")).trim();
  if (!pattern.test(version)) {
    throw new Error(`VERSION must match YYYY.MM.DD-N, got ${version}`);
  }
  return version;
}

async function bumpVersion() {
  const current = await readVersion();
  const today = new Date().toISOString().slice(0, 10).replaceAll("-", ".");
  const [date, number] = current.split("-");
  const next = date === today ? `${today}-${Number(number) + 1}` : `${today}-1`;
  await writeFile(versionPath, `${next}\n`, "utf8");
  return next;
}
