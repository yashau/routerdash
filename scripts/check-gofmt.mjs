import { execFileSync } from "node:child_process";

const output = execFileSync("gofmt", ["-l", "."], { encoding: "utf8" }).trim();
if (output) {
  console.error(`Go files need gofmt:\n${output}`);
  process.exit(1);
}
