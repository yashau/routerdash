export function applySystemTheme() {
  const query = window.matchMedia("(prefers-color-scheme: dark)");
  const apply = () => {
    document.documentElement.classList.toggle("dark", query.matches);
    document.documentElement.style.colorScheme = query.matches ? "dark" : "light";
  };

  apply();
  query.addEventListener("change", apply);
  return () => query.removeEventListener("change", apply);
}
