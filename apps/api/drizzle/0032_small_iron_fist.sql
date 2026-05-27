CREATE TABLE "nook_pool_config" (
	"id" text PRIMARY KEY NOT NULL,
	"workspace_id" text NOT NULL,
	"instance_url" text NOT NULL,
	"secret" text NOT NULL,
	"enabled" boolean DEFAULT false NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "nook_pool_config_workspace_id_unique" UNIQUE("workspace_id")
);
--> statement-breakpoint
ALTER TABLE "project" ADD COLUMN "facile_id" text;--> statement-breakpoint
ALTER TABLE "task" ADD COLUMN "facile_id" text;--> statement-breakpoint
ALTER TABLE "nook_pool_config" ADD CONSTRAINT "nook_pool_config_workspace_id_workspace_id_fk" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspace"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "project" ADD CONSTRAINT "project_facile_id_unique" UNIQUE("facile_id");--> statement-breakpoint
ALTER TABLE "task" ADD CONSTRAINT "task_facile_id_unique" UNIQUE("facile_id");