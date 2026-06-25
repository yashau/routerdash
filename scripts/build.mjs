import { mkdir, readFile } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import { join } from "node:path";

const target = process.argv[2] ?? "linux-amd64";
const root = process.cwd();
const version = (await readFile(join(root, "VERSION"), "utf8")).trim();
const ldflags = `-s -w -X main.version=${version}`;

function run(command, args, options = {}) {
  const executable = process.platform === "win32" && command === "pnpm" ? "cmd.exe" : command;
  const commandArgs =
    process.platform === "win32" && command === "pnpm"
      ? ["/d", "/s", "/c", command, ...args]
      : args;
  const result = spawnSync(executable, commandArgs, {
    cwd: root,
    env: { ...process.env, ...options.env },
    stdio: "inherit",
  });
  if (result.error) {
    console.error(result.error.message);
    process.exit(1);
  }
  if (result.status !== 0) process.exit(result.status ?? 1);
}

await mkdir("dist", { recursive: true });
run("pnpm", ["--dir", "web", "build"]);

if (target === "local") {
  const suffix = process.platform === "win32" ? ".exe" : "";
  run("go", ["build", "-ldflags", ldflags, "-o", join("dist", `routerdash${suffix}`), "."]);
} else if (target === "linux-amd64") {
  await mkdir(join("dist", "routerdash-linux-amd64"), { recursive: true });
  run(
    "go",
    [
      "build",
      "-trimpath",
      "-ldflags",
      ldflags,
      "-o",
      join("dist", "routerdash-linux-amd64", "routerdash"),
      ".",
    ],
    {
      env: { GOOS: "linux", GOARCH: "amd64", CGO_ENABLED: "0" },
    },
  );
} else if (target === "linux-arm64") {
  await mkdir(join("dist", "routerdash-linux-arm64"), { recursive: true });
  run(
    "go",
    [
      "build",
      "-trimpath",
      "-ldflags",
      ldflags,
      "-o",
      join("dist", "routerdash-linux-arm64", "routerdash"),
      ".",
    ],
    {
      env: { GOOS: "linux", GOARCH: "arm64", CGO_ENABLED: "0" },
    },
  );
} else {
  console.error(`Unknown build target: ${target}`);
  process.exit(1);
}
