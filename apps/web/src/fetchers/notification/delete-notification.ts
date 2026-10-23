import { client } from "@opus/libs";

async function deleteNotification(id: string) {
  const response = await client.notification[":id"].$delete({
    param: { id },
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  const data = await response.json();
  return data;
}

export default deleteNotification;
