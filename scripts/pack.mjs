import { createWriteStream } from "node:fs";
import { copyFile, mkdir, readFile, stat } from "node:fs/promises";
import { basename, join } from "node:path";
import { createGzip } from "node:zlib";

const target = process.argv[2] ?? "linux-amd64";
const packageDir = join("dist", `routerdash-${target}`);
const output = join("dist", `routerdash-${target}.tar.gz`);
const version = (await readFile("VERSION", "utf8")).trim();
const versionedOutput = join("dist", `routerdash-${version}-${target}.tar.gz`);

const files = [
  { source: join(packageDir, "routerdash"), name: "routerdash", mode: 0o755 },
  { source: join("deploy", "routerdash.service"), name: "routerdash.service", mode: 0o644 },
  { source: "VERSION", name: "VERSION", mode: 0o644 },
];

await mkdir("dist", { recursive: true });
const gzip = createGzip();
const stream = createWriteStream(output);
gzip.pipe(stream);

for (const file of files) {
  await appendFile(gzip, file.source, file.name, file.mode);
}
gzip.write(Buffer.alloc(1024));
gzip.end();

await new Promise((resolve, reject) => {
  stream.on("finish", resolve);
  stream.on("error", reject);
});

await copyFile(output, versionedOutput);
console.log(`Packed ${output}`);
console.log(`Packed ${versionedOutput}`);

async function appendFile(out, source, name, mode) {
  const info = await stat(source);
  const body = await readFile(source);
  const header = Buffer.alloc(512);
  writeString(header, name || basename(source), 0, 100);
  writeOctal(header, mode, 100, 8);
  writeOctal(header, 0, 108, 8);
  writeOctal(header, 0, 116, 8);
  writeOctal(header, info.size, 124, 12);
  writeOctal(header, Math.floor(info.mtimeMs / 1000), 136, 12);
  header.fill(" ", 148, 156);
  header[156] = "0".charCodeAt(0);
  writeString(header, "ustar", 257, 6);
  writeString(header, "00", 263, 2);
  let checksum = 0;
  for (const byte of header) checksum += byte;
  writeOctal(header, checksum, 148, 8);
  out.write(header);
  out.write(body);
  const padding = (512 - (info.size % 512)) % 512;
  if (padding) out.write(Buffer.alloc(padding));
}

function writeString(buffer, value, offset, length) {
  buffer.write(value.slice(0, length), offset, length, "utf8");
}

function writeOctal(buffer, value, offset, length) {
  const text = value.toString(8).padStart(length - 1, "0") + "\0";
  buffer.write(text, offset, length, "ascii");
}
