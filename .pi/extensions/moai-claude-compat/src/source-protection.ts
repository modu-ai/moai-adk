export function isProjectLocalPath(path: string): boolean {
  return path.startsWith(".pi/");
}
