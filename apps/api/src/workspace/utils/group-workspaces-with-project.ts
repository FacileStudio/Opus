function groupWorkspacesWithProjects(
  workspacesWithProjects: {
    id: string;
    name: string;
    ownerId: string;
    projectId: string | null;
    projectName: string | null;
  }[],
) {
  const groupedWorkspaces = new Map<
    string,
    {
      id: string;
      name: string;
      ownerId: string;
      projects: { projectId: string; projectName: string }[];
    }
  >();

  for (const {
    id,
    name,
    ownerId,
    projectId,
    projectName,
  } of workspacesWithProjects) {
    if (!groupedWorkspaces.has(id)) {
      groupedWorkspaces.set(id, {
        id,
        name,
        ownerId,
        projects: [],
      });
    }

    const workspace = groupedWorkspaces.get(id);
    if (workspace && projectId && projectName) {
      workspace.projects.push({ projectId, projectName });
    }
  }

  return Array.from(groupedWorkspaces.values());
}

export default groupWorkspacesWithProjects;
