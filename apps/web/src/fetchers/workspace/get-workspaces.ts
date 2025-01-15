import { api } from "@kaneo/libs";

const getWorkspaces = async () => {
  const response = await api.workspace.list.get();

  return response;
};

export default getWorkspaces;
