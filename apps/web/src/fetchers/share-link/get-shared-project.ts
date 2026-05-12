import { client } from "@opus/libs";

async function getSharedProject(token: string) {
  const response = await client["shared-project"][":token"].$get({
    param: { token },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json();
}

export default getSharedProject;
