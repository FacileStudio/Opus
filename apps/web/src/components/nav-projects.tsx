import {
  AltArrowRight,
  Folder,
  Forward,
  Link as LinkIcon,
  MenuDots,
  Settings,
  TrashBinTrash,
} from "@solar-icons/react";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "@tanstack/react-router";
import { type MouseEvent, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { MdiPlus as AddCircle } from "@/components/icons/mdi-plus";
import {
  Collapsible,
  CollapsiblePanel,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/menu";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import icons from "@/constants/project-icons";
import createShareLink from "@/fetchers/share-link/create-share-link";
import useDeleteProject from "@/hooks/mutations/project/use-delete-project";
import useGetProjects from "@/hooks/queries/project/use-get-projects";
import useActiveWorkspace from "@/hooks/queries/workspace/use-active-workspace";
import { toast } from "@/lib/toast";
import type { ProjectWithTasks } from "@/types/project";
import type { Workspace } from "@/types/workspace";
import { navGroupLabelClassName, navRowClassName } from "./nav-main";
import CreateProjectModal from "./shared/modals/create-project-modal";
import {
  AlertDialog,
  AlertDialogClose,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";
import { Button } from "./ui/button";

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

function ProjectItem({
  project,
  workspace,
  isMobile,
  isActive,
  onProjectClick,
  onDeleteClick,
}: {
  project: ProjectWithTasks;
  workspace: Workspace;
  isMobile: boolean;
  isActive: boolean;
  onProjectClick: () => void;
  onDeleteClick: () => void;
}) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [contextAnchor, setContextAnchor] = useState<{
    x: number;
    y: number;
  } | null>(null);

  const handleContextMenu = useCallback((e: MouseEvent) => {
    e.preventDefault();
    setContextAnchor({ x: e.clientX, y: e.clientY });
    setOpen(true);
  }, []);

  const ProjectIcon = icons[project.icon as keyof typeof icons] || icons.Layout;

  return (
    <SidebarMenuItem onContextMenu={handleContextMenu}>
      <SidebarMenuButton
        isActive={isActive}
        size="default"
        className={navRowClassName}
        onClick={onProjectClick}
      >
        <ProjectIcon className="size-4.5 shrink-0" />
        <span className="min-w-0 flex-1 truncate">{project.name}</span>
      </SidebarMenuButton>

      <DropdownMenu
        open={open}
        onOpenChange={(nextOpen) => {
          setOpen(nextOpen);
          if (!nextOpen) setContextAnchor(null);
        }}
      >
        <DropdownMenuTrigger
          render={
            <button
              type="button"
              className={`absolute top-3 right-1 flex aspect-square w-5 items-center justify-center rounded-md p-0 outline-hidden ring-sidebar-ring transition-transform focus-visible:ring-2 after:-inset-2 after:absolute md:after:hidden group-data-[collapsible=icon]:hidden group-focus-within/menu-item:opacity-100 group-hover/menu-item:opacity-100 data-[state=open]:opacity-100 md:opacity-0 ${
                isActive
                  ? "text-sidebar-primary-foreground hover:bg-sidebar-primary-foreground/15 hover:text-sidebar-primary-foreground peer-hover/menu-button:text-sidebar-primary-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground peer-hover/menu-button:text-sidebar-accent-foreground"
              }`}
            />
          }
        >
          <MenuDots />
          <span className="sr-only">{t("navigation:sidebar.more")}</span>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          className="w-44 rounded-lg"
          side={contextAnchor ? "bottom" : isMobile ? "bottom" : "right"}
          align={contextAnchor ? "start" : isMobile ? "end" : "start"}
          anchor={
            contextAnchor
              ? makeVirtualAnchor(contextAnchor.x, contextAnchor.y)
              : undefined
          }
        >
          <DropdownMenuItem
            className="h-7 items-start cursor-pointer text-sm"
            onClick={onProjectClick}
          >
            <Folder className="text-muted-foreground" />
            <span>{t("navigation:projectList.viewProject")}</span>
          </DropdownMenuItem>
          <DropdownMenuItem
            className="h-7 items-start cursor-pointer text-sm"
            onClick={() => {
              navigator.clipboard.writeText(
                `${window.location.origin}/dashboard/workspace/${workspace.id}/project/${project.id}`,
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
              try {
                const link = await createShareLink({
                  projectId: project.id,
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
              navigate({
                to: "/dashboard/settings/projects/$projectId/general",
                params: { projectId: project.id },
              });
            }}
          >
            <Settings className="text-muted-foreground" />
            <span>{t("navigation:projectList.projectSettings")}</span>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="h-7 items-start text-destructive cursor-pointer text-sm"
            onClick={onDeleteClick}
          >
            <TrashBinTrash className="text-destructive" />
            <span>{t("navigation:projectList.deleteProject")}</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  );
}

export function NavProjects() {
  const { t } = useTranslation();
  const { isMobile } = useSidebar();
  const { data: workspace } = useActiveWorkspace();
  const { data: projects } = useGetProjects({
    workspaceId: workspace?.id || "",
  });
  const queryClient = useQueryClient();
  const { mutateAsync: deleteProject } = useDeleteProject();
  const navigate = useNavigate();
  const { workspaceId: currentWorkspaceId, projectId: currentProjectId } =
    useParams({
      strict: false,
    });

  const [isCreateProjectModalOpen, setIsCreateProjectModalOpen] =
    useState(false);
  const [isDeleteProjectModalOpen, setIsDeleteProjectModalOpen] =
    useState(false);
  const [projectToDeleteId, setProjectToDeleteID] = useState<string | null>(
    null,
  );

  const isCurrentProject = (projectId: string) => {
    return (
      currentProjectId === projectId && currentWorkspaceId === workspace?.id
    );
  };

  const handleProjectClick = (project: ProjectWithTasks) => {
    navigate({
      to: "/dashboard/workspace/$workspaceId/project/$projectId/board",
      params: {
        workspaceId: workspace?.id || "",
        projectId: project.id,
      },
    });
  };

  if (!workspace) return null;

  return (
    <>
      <Collapsible defaultOpen className="group/collapsible">
        <SidebarGroup className="group-data-[collapsible=icon]:hidden gap-1 p-2">
          <CollapsibleTrigger
            className="data-panel-open:[&_svg]:rotate-90"
            render={
              <SidebarGroupLabel
                className={navGroupLabelClassName}
                render={<button type="button" />}
              />
            }
          >
            <span>{t("navigation:sidebar.projects")}</span>
            <AltArrowRight
              size={14}
              className="transition-transform duration-200"
            />
          </CollapsibleTrigger>
          <CollapsiblePanel>
            <SidebarGroupContent>
              <SidebarMenu className="gap-1">
                {projects?.map((project) => (
                  <ProjectItem
                    key={project.id}
                    project={project}
                    workspace={workspace}
                    isMobile={isMobile}
                    isActive={isCurrentProject(project.id)}
                    onProjectClick={() => handleProjectClick(project)}
                    onDeleteClick={() => {
                      setProjectToDeleteID(project.id);
                      setIsDeleteProjectModalOpen(true);
                    }}
                  />
                ))}

                <SidebarMenuItem className="mt-1">
                  <SidebarMenuButton
                    size="default"
                    className={`${navRowClassName} text-sidebar-foreground/70`}
                    onClick={() => setIsCreateProjectModalOpen(true)}
                  >
                    <AddCircle className="size-4.5 shrink-0" />
                    <span className="min-w-0 flex-1 truncate">
                      {t("navigation:projectList.addProject")}
                    </span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </CollapsiblePanel>
        </SidebarGroup>
      </Collapsible>

      <CreateProjectModal
        open={isCreateProjectModalOpen}
        onClose={() => setIsCreateProjectModalOpen(false)}
      />

      <AlertDialog
        open={isDeleteProjectModalOpen}
        onOpenChange={setIsDeleteProjectModalOpen}
      >
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
                await deleteProject({
                  id: projectToDeleteId || "",
                });
                toast.success(t("navigation:projectList.deletedToast"));
                queryClient.invalidateQueries({
                  queryKey: ["projects"],
                });
                navigate({
                  to: "/dashboard/workspace/$workspaceId",
                  params: {
                    workspaceId: workspace?.id || "",
                  },
                });
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
