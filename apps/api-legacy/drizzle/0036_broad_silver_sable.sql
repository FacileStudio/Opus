CREATE TABLE "nook_pool_outbox" (
	"id" text PRIMARY KEY NOT NULL,
	"seq" bigserial NOT NULL,
	"workspace_id" text NOT NULL,
	"channel" text NOT NULL,
	"payload" text NOT NULL,
	"attempts" integer DEFAULT 0 NOT NULL,
	"last_error" text,
	"created_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "nook_pool_processed_event" (
	"workspace_id" text NOT NULL,
	"idempotency_key" text NOT NULL,
	"processed_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "nook_pool_processed_event_workspace_id_idempotency_key_pk" PRIMARY KEY("workspace_id","idempotency_key")
);
--> statement-breakpoint
ALTER TABLE "project" DROP CONSTRAINT "project_facile_id_unique";--> statement-breakpoint
ALTER TABLE "task" DROP CONSTRAINT "task_facile_id_unique";--> statement-breakpoint
ALTER TABLE "task" DROP CONSTRAINT "task_assignee_id_user_id_fk";
--> statement-breakpoint
ALTER TABLE "nook_pool_outbox" ADD CONSTRAINT "nook_pool_outbox_workspace_id_workspace_id_fk" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspace"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "nook_pool_processed_event" ADD CONSTRAINT "nook_pool_processed_event_workspace_id_workspace_id_fk" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspace"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "nook_pool_outbox_workspace_idx" ON "nook_pool_outbox" USING btree ("workspace_id");--> statement-breakpoint
ALTER TABLE "task" ADD CONSTRAINT "task_actor_id_user_id_fk" FOREIGN KEY ("actor_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE cascade;--> statement-breakpoint
ALTER TABLE "project" ADD CONSTRAINT "project_workspace_facile_id_unique" UNIQUE("workspace_id","facile_id");--> statement-breakpoint
ALTER TABLE "task" ADD CONSTRAINT "task_project_facile_id_unique" UNIQUE("project_id","facile_id");