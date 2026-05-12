import { client } from "@opus/libs";

async function createShareLink(data: {
  projectId: string;
  label?: string;
  expiresAt?: string;
}) {
  const response = await client["share-link"].$post({
    json: data,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json();
}

export default createShareLink;
