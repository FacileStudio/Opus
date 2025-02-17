import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import useGetWorkspaceUsers from "@/hooks/queries/workspace-users/use-get-workspace-users";
import { createFileRoute } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { Shield } from "lucide-react";

export const Route = createFileRoute(
  "/dashboard/teams/$workspaceId/_layout/members",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { workspaceId } = Route.useParams();
  const { data: users } = useGetWorkspaceUsers({ workspaceId });

  return (
    <motion.div
      className="flex-1 p-4 md:p-6"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
    >
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
        <h1 className="text-xl md:text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
          Team Members
        </h1>
      </div>

      <div className="bg-white dark:bg-zinc-900 rounded-lg border border-zinc-200 dark:border-zinc-800 overflow-hidden">
        <div className="overflow-x-auto">
          <div className="md:hidden">
            {users?.map((member) => (
              <motion.div
                key={member.userEmail}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="p-4 border-b border-zinc-200 dark:border-zinc-800 last:border-0"
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback className="bg-indigo-100 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-400 uppercase">
                        {member.userName?.[0]}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <div className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        {member.userName}
                      </div>
                      <div className="text-xs text-zinc-500 dark:text-zinc-400">
                        {member.userEmail}
                      </div>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-400">
                    <Shield className="w-3 h-3" />
                    Admin
                  </span>
                </div>
              </motion.div>
            ))}
          </div>

          <table className="w-full hidden md:table">
            <thead>
              <tr className="border-b border-zinc-200 dark:border-zinc-800">
                <th className="text-left py-3 px-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  Member
                </th>
                <th className="text-left py-3 px-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  Role
                </th>
                <th className="text-left py-3 px-4 text-sm font-medium text-zinc-500 dark:text-zinc-400">
                  Email
                </th>
              </tr>
            </thead>
            <tbody>
              {users?.map((member) => (
                <motion.tr
                  key={member.userEmail}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="border-b border-zinc-200 dark:border-zinc-800 last:border-0"
                >
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-3">
                      <Avatar className="h-6 w-6">
                        <AvatarFallback className="bg-indigo-100 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-400 uppercase text-xs">
                          {member.userName?.[0]}
                        </AvatarFallback>
                      </Avatar>
                      <span className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                        {member.userName}
                      </span>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-2">
                      <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-400">
                        <Shield className="w-3 h-3" />
                        Admin
                      </span>
                    </div>
                  </td>
                  <td className="py-3 px-4 text-sm text-zinc-500 dark:text-zinc-400">
                    {member.userEmail}
                  </td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>

        {users?.length === 0 && (
          <div className="py-12 text-center">
            <p className="text-zinc-500 dark:text-zinc-400">
              No team members found
            </p>
          </div>
        )}
      </div>
    </motion.div>
  );
}
