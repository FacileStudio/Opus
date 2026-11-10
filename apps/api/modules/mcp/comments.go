package mcp

import (
	"context"

	"github.com/FacileStudio/Opus/apps/api/internal/wireid"
	"github.com/FacileStudio/Opus/apps/api/modules/activity"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// commentPage is how many comments one opus_read_task call returns.
const commentPage = 20

// commentKind is the activity type of a comment.
const commentKind = "comment"

// attachComments adds one page of the task's comments to out, newest first,
// from offset.
func (t *toolset) attachComments(ctx context.Context, userID int64, out *taskOutput, offset int) error {
	activities, err := t.domain.Activities.Calls().List(ctx, userID, out.TaskID)
	if err != nil {
		return mcpkit.Failure(err, hintTask)
	}
	comments := make([]activity.ActivityResponse, 0, len(activities))
	for _, a := range activities {
		if a.Type == commentKind && a.Content != nil {
			comments = append(comments, a)
		}
	}
	if offset > len(comments) {
		return mcpkit.ErrCursor
	}
	rest := comments[offset:min(offset+commentPage, len(comments))]
	authors := t.authors(ctx, rest)
	rendered := make([]comment, len(rest))
	for i, a := range rest {
		rendered[i] = render(a, authors)
	}
	keep := mcpkit.FitCount(len(rendered), func(n int) int {
		probe := *out
		probe.Comments, probe.CommentsNextCursor = rendered[:n], "next-cursor-placeholder"
		return mcpkit.EncodedSize(probe)
	})
	out.Comments, out.CommentsTotal = rendered[:keep], len(comments)
	if offset+keep < len(comments) {
		out.CommentsNextCursor = mcpkit.EncodeCursor(position{Offset: offset + keep})
	}
	return nil
}

// authors loads the people who wrote comments.
func (t *toolset) authors(ctx context.Context, list []activity.ActivityResponse) map[int64]*person {
	ids := make([]int64, 0, len(list))
	for _, a := range list {
		if a.UserID != nil {
			ids = append(ids, *a.UserID)
		}
	}
	out := map[int64]*person{}
	if len(ids) == 0 {
		return out
	}
	var users []schemas.User
	t.domain.DB.WithContext(ctx).Select("id", "name").Where("id IN ?", ids).Find(&users)
	for _, u := range users {
		out[u.ID] = &person{UserID: u.ID, Name: u.Name}
	}
	return out
}

// render turns a comment activity into a comment.
func render(a activity.ActivityResponse, authors map[int64]*person) comment {
	text, cut := mcpkit.Clip(*a.Content, commentLimit)
	c := comment{CommentID: a.ID, Text: text, Truncated: cut, CreatedAt: stamp(a.CreatedAt), Edited: a.UpdatedAt.After(a.CreatedAt)}
	if a.UserID != nil {
		c.Author = authors[*a.UserID]
	}
	return c
}

// commentTask is opus_comment_task.
func (t *toolset) commentTask(ctx context.Context, req *sdk.CallToolRequest, in commentInput) (*sdk.CallToolResult, commentOutput, error) {
	ctx, userID, err := t.caller(ctx, req)
	if err != nil {
		return nil, commentOutput{}, err
	}
	posted, err := t.domain.Activities.Calls().Comment(ctx, userID, &activity.CommentActivityRequest{
		TaskID: wireid.ID(in.TaskID), Comment: in.Comment,
	})
	if err != nil {
		return nil, commentOutput{}, mcpkit.Failure(err, hintTask)
	}
	return nil, commentOutput{
		Status: "posted", Message: "The comment is on the task.", TaskID: in.TaskID,
		Comment: render(*posted, t.authors(ctx, []activity.ActivityResponse{*posted})),
	}, nil
}
