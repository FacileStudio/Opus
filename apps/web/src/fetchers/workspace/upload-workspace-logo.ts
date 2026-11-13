import { apiFetch, getApiUrl } from "@/fetchers/get-api-url";

async function uploadWorkspaceLogo(workspaceId: string, file: File) {
  const formData = new FormData();
  formData.append("logo", file);

  const response = await apiFetch(getApiUrl(`/workspace/${workspaceId}/logo`), {
    method: "POST",
    body: formData,
    credentials: "include",
    headers: {
      "X-Facile-CSRF": "1",
    },
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ message: "Upload failed" }));
    throw new Error((error as { message?: string }).message || "Upload failed");
  }

  return response.json() as Promise<{ url: string; message: string }>;
}

async function deleteWorkspaceLogo(workspaceId: string) {
  const response = await apiFetch(getApiUrl(`/workspace/${workspaceId}/logo`), {
    method: "DELETE",
    credentials: "include",
    headers: {
      "X-Facile-CSRF": "1",
    },
  });

  if (!response.ok) {
    throw new Error("Failed to delete logo");
  }

  return response.json() as Promise<{ message: string }>;
}

export { deleteWorkspaceLogo, uploadWorkspaceLogo };
