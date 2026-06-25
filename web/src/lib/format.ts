export function bytes(value: number) {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit += 1;
  }
  return `${size.toFixed(size >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`;
}

export function rate(value: number) {
  return `${bytes(value)}/s`;
}

export function pct(value: number) {
  return `${value.toFixed(1)}%`;
}
