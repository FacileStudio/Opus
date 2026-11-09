// Package events publishes outbound domain events to the Antenne event bus.
package events

import "github.com/FacileStudio/enveloppe/go"

const (
	// ChannelTaskCreated carries task creation events.
	ChannelTaskCreated = "task.created"
	// ChannelTaskUpdated carries task update events.
	ChannelTaskUpdated = "task.updated"
	// ChannelTaskDeleted carries task deletion events.
	ChannelTaskDeleted = "task.deleted"
	// ChannelTaskStatusChanged carries task workflow column or status transitions.
	ChannelTaskStatusChanged = "task.status_changed"
	// ChannelTaskActorChanged carries task assignee changes.
	ChannelTaskActorChanged = "task.actor_changed"

	// ChannelProjectCreated carries project creation events.
	ChannelProjectCreated = "project.created"
	// ChannelProjectUpdated carries project update events.
	ChannelProjectUpdated = "project.updated"
	// ChannelProjectDeleted carries project deletion events.
	ChannelProjectDeleted = "project.deleted"

	// ChannelTimeEntryCreated carries Sablier time entry creation events.
	ChannelTimeEntryCreated = "time_entry.created"
	// ChannelTimeEntryUpdated carries Sablier time entry update events.
	ChannelTimeEntryUpdated = "time_entry.updated"
	// ChannelTimeEntryDeleted carries Sablier time entry deletion events.
	ChannelTimeEntryDeleted = "time_entry.deleted"

	// ChannelCommentCreated carries new task comment events.
	ChannelCommentCreated = "comment.created"

	// ChannelWorkspaceCreated carries workspace creation events.
	ChannelWorkspaceCreated = "workspace.created"

	// ChannelGitPullRequest carries pull request events forwarded by Antenne.
	ChannelGitPullRequest = "git.pull_request"
	// ChannelGitPush carries push events forwarded by Antenne.
	ChannelGitPush = "git.push"

	// ObjectComment is the wire object type for comment events.
	ObjectComment enveloppe.ObjectType = "comment"
	// ObjectWorkspace is the wire object type for workspace events.
	ObjectWorkspace enveloppe.ObjectType = "workspace"
)

// AllChannels returns the list of all event channels Opus emits.
func AllChannels() []string {
	return []string{
		ChannelTaskCreated,
		ChannelTaskUpdated,
		ChannelTaskDeleted,
		ChannelTaskStatusChanged,
		ChannelTaskActorChanged,
		ChannelProjectCreated,
		ChannelProjectUpdated,
		ChannelProjectDeleted,
		ChannelCommentCreated,
		ChannelWorkspaceCreated,
	}
}

// SyncChannels returns the channels Opus listens to for Sablier project, task
// and time entry sync.
func SyncChannels() []string {
	return []string{
		ChannelProjectCreated,
		ChannelProjectUpdated,
		ChannelProjectDeleted,
		ChannelTaskCreated,
		ChannelTaskUpdated,
		ChannelTaskDeleted,
		ChannelTimeEntryCreated,
		ChannelTimeEntryUpdated,
		ChannelTimeEntryDeleted,
	}
}

// GitChannels returns the channels Opus listens to for git forge events.
func GitChannels() []string {
	return []string{
		ChannelGitPullRequest,
		ChannelGitPush,
	}
}

// InboundChannels returns all channels Opus registers with the Antenne pool.
func InboundChannels() []string {
	return append(SyncChannels(), GitChannels()...)
}
