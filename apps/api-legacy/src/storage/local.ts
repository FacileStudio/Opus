import { existsSync } from "node:fs";
import { mkdir, readFile, unlink, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";

const ALLOWED_IMAGE_TYPES = new Set([
  "image/png",
  "image/jpeg",
  "image/jpg",
  "image/webp",
  "image/svg+xml",
  "image/gif",
]);

const MAX_LOGO_BYTES = 2 * 1024 * 1024;

function getUploadDir(): string {
  return resolve(
    process.env.UPLOAD_DIR || join(process.cwd(), "data", "uploads"),
  );
}

export function getWorkspaceLogoDir(workspaceId: string): string {
  return join(getUploadDir(), "workspaces", workspaceId);
}

export function getWorkspaceLogoPath(
  workspaceId: string,
  extension: string,
): string {
  return join(getWorkspaceLogoDir(workspaceId), `logo.${extension}`);
}

export function extensionFromMime(mime: string): string {
  const map: Record<string, string> = {
    "image/png": "png",
    "image/jpeg": "jpg",
    "image/jpg": "jpg",
    "image/webp": "webp",
    "image/svg+xml": "svg",
    "image/gif": "gif",
  };
  return map[mime] || "png";
}

export function validateLogoUpload(contentType: string, size: number): void {
  if (!ALLOWED_IMAGE_TYPES.has(contentType.toLowerCase())) {
    throw new Error(
      `Invalid image type: ${contentType}. Allowed: png, jpeg, webp, svg, gif`,
    );
  }

  if (size > MAX_LOGO_BYTES) {
    throw new Error(
      `File too large (${Math.round(size / 1024)}KB). Maximum: ${MAX_LOGO_BYTES / (1024 * 1024)}MB`,
    );
  }
}

export async function saveWorkspaceLogo(
  workspaceId: string,
  buffer: Buffer,
  extension: string,
): Promise<string> {
  const logoPath = getWorkspaceLogoPath(workspaceId, extension);
  const dir = dirname(logoPath);

  await mkdir(dir, { recursive: true });

  await removeExistingLogos(workspaceId, extension);

  await writeFile(logoPath, buffer);
  return logoPath;
}

async function removeExistingLogos(
  workspaceId: string,
  keepExtension: string,
): Promise<void> {
  const extensions = ["png", "jpg", "webp", "svg", "gif"];
  const dir = getWorkspaceLogoDir(workspaceId);

  for (const ext of extensions) {
    if (ext === keepExtension) continue;
    const path = join(dir, `logo.${ext}`);
    if (existsSync(path)) {
      await unlink(path);
    }
  }
}

export async function readWorkspaceLogo(
  workspaceId: string,
): Promise<{ buffer: Buffer; extension: string } | null> {
  const extensions = ["webp", "png", "jpg", "svg", "gif"];

  for (const ext of extensions) {
    const path = getWorkspaceLogoPath(workspaceId, ext);
    if (existsSync(path)) {
      const buffer = await readFile(path);
      return { buffer, extension: ext };
    }
  }

  return null;
}

export async function deleteWorkspaceLogo(
  workspaceId: string,
): Promise<boolean> {
  const extensions = ["png", "jpg", "webp", "svg", "gif"];
  let deleted = false;

  for (const ext of extensions) {
    const path = getWorkspaceLogoPath(workspaceId, ext);
    if (existsSync(path)) {
      await unlink(path);
      deleted = true;
    }
  }

  return deleted;
}
