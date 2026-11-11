// Package events publishes outbound domain events to the Antenne event bus.
package events

import (
	"github.com/FacileStudio/antenne-client/go/catalog"
	envelope "github.com/FacileStudio/antenne-client/go/envelope"
)

const (
	// ChannelTaskCreated carries task creation events.
	ChannelTaskCreated = string(catalog.TaskCreated)
	// ChannelTaskUpdated carries task update events.
	ChannelTaskUpdated = string(catalog.TaskUpdated)
	// ChannelTaskDeleted carries task deletion events.
	ChannelTaskDeleted = string(catalog.TaskDeleted)
	// ChannelTaskStatusChanged carries task workflow column or status transitions.
	ChannelTaskStatusChanged = string(catalog.TaskStatusChanged)
	// ChannelTaskActorChanged carries task assignee changes.
	ChannelTaskActorChanged = string(catalog.TaskActorChanged)

	// ChannelProjectCreated carries project creation events.
	ChannelProjectCreated = string(catalog.ProjectCreated)
	// ChannelProjectUpdated carries project update events.
	ChannelProjectUpdated = string(catalog.ProjectUpdated)
	// ChannelProjectDeleted carries project deletion events.
	ChannelProjectDeleted = string(catalog.ProjectDeleted)

	// ChannelTimeEntryCreated carries Sablier time entry creation events.
	ChannelTimeEntryCreated = string(catalog.TimeEntryCreated)
	// ChannelTimeEntryUpdated carries Sablier time entry update events.
	ChannelTimeEntryUpdated = string(catalog.TimeEntryUpdated)
	// ChannelTimeEntryDeleted carries Sablier time entry deletion events.
	ChannelTimeEntryDeleted = string(catalog.TimeEntryDeleted)

	// ChannelCommentCreated carries new task comment events.
	ChannelCommentCreated = string(catalog.CommentCreated)

	// ChannelWorkspaceCreated carries workspace creation events.
	ChannelWorkspaceCreated = string(catalog.WorkspaceCreated)

	// ChannelGitPullRequest carries pull request events forwarded by Antenne.
	ChannelGitPullRequest = string(catalog.GitPullRequest)
	// ChannelGitPush carries push events forwarded by Antenne.
	ChannelGitPush = string(catalog.GitPush)

	// ObjectComment is the wire object type for comment events.
	ObjectComment envelope.ObjectType = "comment"
	// ObjectWorkspace is the wire object type for workspace events.
	ObjectWorkspace envelope.ObjectType = "workspace"
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
