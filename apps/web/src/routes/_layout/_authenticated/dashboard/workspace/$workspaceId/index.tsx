import {
  Folder,
  Forward,
  Widget as LayoutGrid,
  Link as LinkIcon,
  List,
  Settings,
  TrashBinTrash,
  Widget3,
} from "@solar-icons/react";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type MouseEvent, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import WorkspaceLayout from "@/components/common/workspace-layout";
import { MdiPlus as Plus } from "@/components/icons/mdi-plus";
import PageTitle from "@/components/page-title";
import CreateProjectModal from "@/components/shared/modals/create-project-modal";
import {
  AlertDialog,
  AlertDialogClose,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/menu";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import icons from "@/constants/project-icons";
import { shortcuts } from "@/constants/shortcuts";
import createShareLink from "@/fetchers/share-link/create-share-link";
import useDeleteProject from "@/hooks/mutations/project/use-delete-project";
import useGetProjects from "@/hooks/queries/project/use-get-projects";
import { useRegisterShortcuts } from "@/hooks/use-keyboard-shortcuts";
import { formatDateMedium } from "@/lib/format";
import { toast } from "@/lib/toast";
import { useUserPreferencesStore } from "@/store/user-preferences";

function makeVirtualAnchor(x: number, y: number) {
  const rect = {
    x,
    y,
    width: 0,
    height: 0,
    top: y,
    right: x,
    bottom: y,
    left: x,
    toJSON: () => rect,
  };
  return { getBoundingClientRect: () => rect };
}

export const Route = createFileRoute(
  "/_layout/_authenticated/dashboard/workspace/$workspaceId/",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { t } = useTranslation();
  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false);
  const { workspaceId } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: projects, isLoading } = useGetProjects({
    workspaceId,
  });
  const { mutateAsync: deleteProject } = useDeleteProject();
  const projectListView = useUserPreferencesStore((s) => s.projectListView);
  const setProjectListView = useUserPreferencesStore(
    (s) => s.setProjectListView,
  );

  const [contextMenu, setContextMenu] = useState<{
    projectId: string;
    x: number;
    y: number;
  } | null>(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [projectToDeleteId, setProjectToDeleteId] = useState<string | null>(
    null,
  );

  const handleContextMenu = useCallback((e: MouseEvent, projectId: string) => {
    e.preventDefault();
    setContextMenu({ projectId, x: e.clientX, y: e.clientY });
  }, []);

  const handleCreateProject = () => {
    setIsCreateProjectOpen(true);
  };

  useRegisterShortcuts({
    sequentialShortcuts: {
      [shortcuts.project.prefix]: {
        [shortcuts.project.create]: handleCreateProject,
      },
    },
  });

  const handleProjectClick = (projectId: string) => {
    navigate({
      to: "/dashboard/workspace/$workspaceId/project/$projectId/board",
      params: { workspaceId, projectId },
    });
  };

  if (isLoading) {
    return (
      <>
        <PageTitle title={t("workspace:projects.pageTitle")} />
        <WorkspaceLayout
          title={t("workspace:projects.pageTitle")}
          headerActions={
            <Button
              variant="outline"
              size="xs"
              onClick={handleCreateProject}
              className="gap-1"
            >
              <Plus className="w-3 h-3" />
              {t("workspace:projects.createProject")}
            </Button>
          }
        >
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.title")}
                </TableHead>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.progress")}
                </TableHead>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.targetDate")}
                </TableHead>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.status")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {[1, 2, 3].map((i) => (
                <TableRow key={i}>
                  <TableCell className="py-3">
                    <div className="flex items-center gap-3">
                      <Skeleton className="h-5 w-5" />
                      <Skeleton className="h-4 w-24" />
                    </div>
                  </TableCell>
                  <TableCell className="py-3">
                    <Skeleton className="h-2 w-20" />
                  </TableCell>
                  <TableCell className="py-3">
                    <Skeleton className="h-4 w-20" />
                  </TableCell>
                  <TableCell className="py-3">
                    <Skeleton className="h-5 w-16" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </WorkspaceLayout>
      </>
    );
  }

  if (!projects || projects.length === 0) {
    return (
      <>
        <PageTitle title={t("workspace:projects.pageTitle")} />
        <WorkspaceLayout
          title={t("workspace:projects.pageTitle")}
          headerActions={
            <Button
              variant="outline"
              size="xs"
              onClick={handleCreateProject}
              className="gap-1"
            >
              <Plus className="w-3 h-3" />
              {t("workspace:projects.createProject")}
            </Button>
          }
        >
          <Empty className="min-h-[60vh]">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <LayoutGrid />
              </EmptyMedia>
              <EmptyTitle>{t("workspace:projects.emptyTitle")}</EmptyTitle>
              <EmptyDescription>
                {t("workspace:projects.emptyDescription")}
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button onClick={handleCreateProject}>
                <Plus />
                {t("workspace:projects.createProject")}
              </Button>
            </EmptyContent>
          </Empty>
        </WorkspaceLayout>

        <CreateProjectModal
          open={isCreateProjectOpen}
          onClose={() => setIsCreateProjectOpen(false)}
        />
      </>
    );
  }

  return (
    <>
      <PageTitle title={t("workspace:projects.pageTitle")} />
      <WorkspaceLayout
        title={t("workspace:projects.pageTitle")}
        headerActions={
          <div className="flex items-center gap-2">
            <div className="inline-flex items-center gap-1 rounded-lg bg-muted/50 p-0.5">
              <button
                type="button"
                className={`inline-flex h-6 items-center gap-1 rounded-md px-2 text-xs font-medium transition-colors ${
                  projectListView === "card"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                }`}
                onClick={() => setProjectListView("card")}
              >
                <Widget3 size={12} />
              </button>
              <button
                type="button"
                className={`inline-flex h-6 items-center gap-1 rounded-md px-2 text-xs font-medium transition-colors ${
                  projectListView === "list"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                }`}
                onClick={() => setProjectListView("list")}
              >
                <List size={12} />
              </button>
            </div>
            <Button
              variant="outline"
              size="xs"
              onClick={handleCreateProject}
              className="gap-1"
            >
              <Plus className="w-3 h-3" />
              {t("workspace:projects.createProject")}
            </Button>
          </div>
        }
      >
        {projectListView === "card" ? (
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {projects?.map((project) => {
              if (!project || !project.id || !project.statistics) return null;

              const IconComponent =
                icons[project.icon as keyof typeof icons] || icons.Layout;

              const getStatusText = () => {
                if (project.statistics.totalTasks === 0)
                  return t("workspace:projects.projectStatus.notStarted");
                if (project.statistics.completionPercentage === 100)
                  return t("workspace:projects.projectStatus.complete");
                return t("workspace:projects.projectStatus.inProgress");
              };

              const getStatusVariant = () => {
                if (project.statistics.totalTasks === 0) return "secondary";
                if (project.statistics.completionPercentage === 100)
                  return "default";
                return "outline";
              };

              return (
                <button
                  type="button"
                  key={project.id}
                  className="flex cursor-pointer flex-col gap-4 rounded-xl border border-border bg-card py-4 text-left text-sm ring-1 ring-foreground/5 transition-colors hover:bg-muted/40"
                  onClick={() => handleProjectClick(project.id)}
                  onContextMenu={(e) => handleContextMenu(e, project.id)}
                >
                  <div className="flex items-start justify-between gap-2 px-4">
                    <div className="flex min-w-0 flex-1 items-start gap-2.5">
                      <IconComponent className="mt-0.5 h-5 w-5 shrink-0 text-muted-foreground" />
                      <div className="min-w-0">
                        <p className="truncate text-base font-medium">
                          {project.name}
                        </p>
                        <p className="text-sm text-muted-foreground">
                          {project.description || "—"}
                        </p>
                      </div>
                    </div>
                    <Badge variant={getStatusVariant()} className="shrink-0">
                      {getStatusText()}
                    </Badge>
                  </div>
                  <div className="space-y-2 px-4">
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span>
                        {project.statistics.totalTasks}{" "}
                        {project.statistics.totalTasks === 1 ? "task" : "tasks"}
                      </span>
                      <span>{project.statistics.completionPercentage}%</span>
                    </div>
                    <Progress
                      value={project.statistics.completionPercentage}
                      className="h-1.5"
                    />
                  </div>
                </button>
              );
            })}
          </div>
        ) : (
          <Table>
            <TableHeader className="p-4">
              <TableRow>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.title")}
                </TableHead>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.progress")}
                </TableHead>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.dueDate")}
                </TableHead>
                <TableHead className="text-foreground font-medium">
                  {t("workspace:projects.status")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {projects?.map((project) => {
                if (!project || !project.id || !project.statistics) return null;

                const IconComponent =
                  icons[project.icon as keyof typeof icons] || icons.Layout;

                const getStatusText = () => {
                  if (project.statistics.totalTasks === 0)
                    return t("workspace:projects.projectStatus.notStarted");
                  if (project.statistics.completionPercentage === 100)
                    return t("workspace:projects.projectStatus.complete");
                  return t("workspace:projects.projectStatus.inProgress");
                };

                const getStatusVariant = () => {
                  if (project.statistics.totalTasks === 0) return "secondary";
                  if (project.statistics.completionPercentage === 100)
                    return "default";
                  return "outline";
                };

                return (
                  <TableRow
                    key={project.id}
                    className="cursor-pointer"
                    onClick={() => handleProjectClick(project.id)}
                    onContextMenu={(e) => handleContextMenu(e, project.id)}
                  >
                    <TableCell className="py-3">
                      <div className="flex items-center gap-3">
                        <IconComponent className="w-5 h-5 text-muted-foreground" />
                        <span className="font-medium">{project.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="py-3">
                      <div className="flex items-center gap-2">
                        <Progress
                          value={project.statistics.completionPercentage}
                          className="w-16 h-2"
                        />
                        <span className="text-sm text-muted-foreground">
                          {project.statistics.completionPercentage}%
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="py-3">
                      <span className="text-sm text-muted-foreground">
                        {project.statistics.dueDate
                          ? formatDateMedium(project.statistics.dueDate)
                          : t("workspace:projects.noDueDate")}
                      </span>
                    </TableCell>
                    <TableCell className="py-3">
                      <Badge variant={getStatusVariant()}>
                        {getStatusText()}
                      </Badge>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}

        <DropdownMenu
          open={contextMenu !== null}
          onOpenChange={(open) => {
            if (!open) setContextMenu(null);
          }}
        >
          <DropdownMenuContent
            className="w-44 rounded-lg"
            side="bottom"
            align="start"
            anchor={
              contextMenu
                ? makeVirtualAnchor(contextMenu.x, contextMenu.y)
                : undefined
            }
          >
            <DropdownMenuItem
              className="h-7 items-start cursor-pointer text-sm"
              onClick={() => {
                if (contextMenu) handleProjectClick(contextMenu.projectId);
              }}
            >
              <Folder className="text-muted-foreground" />
              <span>{t("navigation:projectList.viewProject")}</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              className="h-7 items-start cursor-pointer text-sm"
              onClick={() => {
                if (!contextMenu) return;
                navigator.clipboard.writeText(
                  `${window.location.origin}/dashboard/workspace/${workspaceId}/project/${contextMenu.projectId}`,
                );
                toast.success(t("navigation:projectList.linkCopied"));
              }}
            >
              <Forward className="text-muted-foreground" />
              <span>{t("navigation:projectList.shareProject")}</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              className="h-7 items-start cursor-pointer text-sm"
              onClick={async () => {
                if (!contextMenu) return;
                try {
                  const link = await createShareLink({
                    projectId: contextMenu.projectId,
                  });
                  await navigator.clipboard.writeText(
                    `${window.location.origin}/shared/${link.token}`,
                  );
                  toast.success(t("navigation:projectList.shareLinkCreated"));
                } catch {
                  toast.error(t("navigation:projectList.shareLinkError"));
                }
              }}
            >
              <LinkIcon className="text-muted-foreground" />
              <span>{t("navigation:projectList.createShareLink")}</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              className="h-7 items-start cursor-pointer text-sm"
              onClick={() => {
                if (!contextMenu) return;
                navigate({
                  to: "/dashboard/settings/projects/$projectId/general",
                  params: { projectId: contextMenu.projectId },
                });
              }}
            >
              <Settings className="text-muted-foreground" />
              <span>{t("navigation:projectList.projectSettings")}</span>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="h-7 items-start text-destructive cursor-pointer text-sm"
              onClick={() => {
                if (!contextMenu) return;
                setProjectToDeleteId(contextMenu.projectId);
                setIsDeleteModalOpen(true);
              }}
            >
              <TrashBinTrash className="text-destructive" />
              <span>{t("navigation:projectList.deleteProject")}</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </WorkspaceLayout>

      <CreateProjectModal
        open={isCreateProjectOpen}
        onClose={() => setIsCreateProjectOpen(false)}
      />

      <AlertDialog open={isDeleteModalOpen} onOpenChange={setIsDeleteModalOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("navigation:projectList.deleteConfirmTitle")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("navigation:projectList.deleteConfirmDescription")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogClose>
              <Button variant="outline" size="sm">
                {t("common:actions.cancel")}
              </Button>
            </AlertDialogClose>
            <AlertDialogClose
              onClick={async () => {
                await deleteProject({ id: projectToDeleteId || "" });
                toast.success(t("navigation:projectList.deletedToast"));
                queryClient.invalidateQueries({ queryKey: ["projects"] });
              }}
            >
              <Button variant="destructive" size="sm">
                {t("navigation:projectList.deleteProject")}
              </Button>
            </AlertDialogClose>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
