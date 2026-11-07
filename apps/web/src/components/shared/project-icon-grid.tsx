import { useState } from "react";
import { Input } from "@/components/ui/input";
import icons from "@/constants/project-icons";
import { cn } from "@/lib/cn";

type ProjectIconGridProps = {
  value: string;
  placeholder: string;
  onPick: (icon: string) => void;
};

/**
 * The searchable grid of project icons shown inside an icon popover. Cells are plain
 * buttons of one fixed size and the grid fills its width, so nothing overlaps at any
 * popover width or on touch screens.
 */
export default function ProjectIconGrid({
  value,
  placeholder,
  onPick,
}: ProjectIconGridProps) {
  const [search, setSearch] = useState("");
  const query = search.trim().toLowerCase();
  const matches = Object.entries(icons).filter(([name]) =>
    name.toLowerCase().includes(query),
  );

  return (
    <div className="flex flex-col gap-2 p-1">
      <Input
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        placeholder={placeholder}
        className="h-8 text-xs"
      />
      <div className="max-h-[280px] overflow-y-auto overscroll-contain">
        <div className="grid grid-cols-[repeat(auto-fill,2.25rem)] justify-between gap-1">
          {matches.map(([name, Icon]) => (
            <button
              key={name}
              type="button"
              title={name}
              aria-label={name}
              aria-pressed={value === name}
              onClick={() => onPick(name)}
              className={cn(
                "flex size-9 items-center justify-center rounded-md text-muted-foreground outline-none transition-colors hover:bg-accent hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring",
                value === name &&
                  "bg-sidebar-accent text-sidebar-accent-foreground",
              )}
            >
              <Icon className="size-4" />
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
