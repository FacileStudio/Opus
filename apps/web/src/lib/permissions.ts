// Permission statements and role definitions, replacing better-auth's access
// plugin. The server owns enforcement; this module is the client-side mirror
// the UI reads to decide which controls to render, and the single place the
// role→resources mapping lives.

export type PermissionResource = "project" | "task" | "workspace" | "team";

export const statement = {
  project: ["create", "read", "update", "delete", "share"],
  task: ["create", "read", "update", "delete", "assign"],
  workspace: ["read", "update", "delete", "manage_settings"],
  team: ["invite", "remove", "manage_roles"],
} as const;

export type Role = "owner" | "admin" | "member";

type RuleMap = Record<PermissionResource, readonly string[]>;

const memberRules: RuleMap = {
  project: ["create", "read"],
  task: ["create", "read", "update"],
  workspace: ["read"],
  team: [],
};

const adminRules: RuleMap = {
  project: ["create", "read", "update", "delete", "share"],
  task: ["create", "read", "update", "delete", "assign"],
  workspace: ["read", "update", "manage_settings"],
  team: ["invite", "remove", "manage_roles"],
};

const ownerRules: RuleMap = {
  project: ["create", "read", "update", "delete", "share"],
  task: ["create", "read", "update", "delete", "assign"],
  workspace: ["read", "update", "delete", "manage_settings"],
  team: ["invite", "remove", "manage_roles"],
};

export const ac: Record<Role, { permissions: RuleMap }> = {
  member: { permissions: memberRules },
  admin: { permissions: adminRules },
  owner: { permissions: ownerRules },
};

export const member = ac.member;
export const admin = ac.admin;
export const owner = ac.owner;

// roleHas answers whether a role may perform action on resource. It is the
// only client-side gate the UI consults (via auth-client.checkRolePermission);
// the server re-enforces every one of these on the wire.
export function roleHas(role: Role, resource: string, action: string): boolean {
  const rules = ac[role]?.permissions;
  return Boolean(rules?.[resource as PermissionResource]?.includes(action));
}
