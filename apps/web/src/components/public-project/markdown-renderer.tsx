import CommentEditor from "@/components/activity/comment-editor";

type MarkdownRendererProps = {
  content: string;
};

export function MarkdownRenderer({ content }: MarkdownRendererProps) {
  return (
    <CommentEditor
      value={content}
      readOnly
      showBubbleMenu={false}
      proseClassName="opus-tiptap-prose"
      contentClassName="opus-tiptap-content"
      className="[&_.opus-tiptap-content_.ProseMirror]:max-h-none [&_.opus-tiptap-content_.ProseMirror]:overflow-visible [&_.opus-tiptap-content_.ProseMirror]:px-0 [&_.opus-tiptap-content_.ProseMirror]:py-0"
    />
  );
}
