import { client } from "@opus/libs";

async function updateShareLink(
  id: string,
  data: {
    label?: string | null;
    expiresAt?: string | null;
    isActive?: boolean;
  },
) {
  const response = await client["share-link"][":id"].$put({
    param: { id },
    json: data,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json();
}

export default updateShareLink;
