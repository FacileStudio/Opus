import { client } from "@opus/libs";

async function getShareLinks(projectId: string) {
  const response = await client["share-link"].$get({
    query: { projectId },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json();
}

export default getShareLinks;
