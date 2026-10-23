import CircularProgress from "@/components/ui/circular-progress";

type SubtaskProgressProps = {
  done?: number;
  total?: number;
};

function SubtaskProgress({ done = 0, total = 0 }: SubtaskProgressProps) {
  if (total <= 0) return null;

  return (
    <span className="flex shrink-0 items-center gap-1 text-[10px] text-muted-foreground">
      <CircularProgress completed={done} total={total} size={12} />
      {done}/{total}
    </span>
  );
}

export default SubtaskProgress;
