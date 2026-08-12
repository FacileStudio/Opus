// Permission statements and role definitions, replacing better-auth's access
// plugin. The server owns enforcement; this module is the client-side mirror
// the UI reads to decide which controls to render, and the single place the
// role→resources mapping lives.

export type PermissionResource = "project" | "task" | "workspace" | "team";

const statement = {
  project: ["create", "read", "update", "delete", "share"],
  task: ["create", "read", "update", "delete", "assign"],
  workspace: ["read", "update", "delete", "manage_settings"],
  team: ["invite", "remove", "manage_roles"],
} as const;

export type PermissionStatement = {
  [R in PermissionResource]: readonly string[];
};

type Role = "owner" | "admin" | "member";

// NonNullable<Record<string, readonly string[]>> shape for the permission API.
type RuleMap = Record<PermissionResource, readonly string[] | undefined>;

function buildRole(project: readonly string[], task: readonly string[], workspace: readonly string[], team: readonly string[]): { permissions: RuleMap } {
  return {
    permissions: {
      project,
      task,
      workspace,
      team,
    },
  };
}

const MemberAC = buildRole(["create", "read"], ["create", "read", "update"], ["read"], []);
const AdminAC = buildRole(
  ["create", "read", "update", "delete", "share"],
  ["create", "read", "update", "delete", "assign"],
  ["read", "update", "manage_settings"],
  ["invite", "remove", "manage_roles"],
);
const OwnerAC = buildRole(
  ["create", "read", "update", "delete", "share"],
  ["create", "read", "update", "delete", "assign"],
  ["read", "update", "delete", "manage_settings"],
  ["invite", "remove", "manage_roles"],
);

const ac = {
  member: MemberAC,
  admin: AdminAC,
  owner: OwnerAC,
} as const;

const member = MemberAC;
const admin = AdminAC;
const owner = OwnerAC;

export { ac, admin, member, owner, statement };
export type { Role };