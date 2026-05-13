import * as SolarIcons from "@solar-icons/react";
import type { ComponentType, SVGProps } from "react";

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>;

const EXCLUDED = new Set(["IconBase", "SolarProvider", "useSolar", "default"]);

function isTypoDuplicate(name: string): boolean {
  if (name.includes("Minimlistic") || name.includes("Minimlistic")) return true;
  if (name.includes("Minimlistik")) return true;
  if (
    name.startsWith("Clound") ||
    name.startsWith("Pallete") ||
    name.startsWith("Siderbar")
  )
    return true;
  if (name.startsWith("Plaaylist")) return true;
  if (name.startsWith("Magnifer") && !name.startsWith("Magnifier")) return true;
  if (name.startsWith("MinimalisticMagnifer")) return true;
  if (name.startsWith("MinimlisticMagnifier")) return true;
  if (name.startsWith("RoundedMagnifer")) return true;
  if (
    name === "Recive" ||
    name === "ReciveSquare" ||
    name === "ReciveTwiceSquare"
  )
    return true;
  if (name === "CardRecive") return true;
  if (name === "ColourTuneing") return true;
  if (name === "Condicioner" || name === "Condicioner2") return true;
  if (name === "MaskHapply") return true;
  if (
    name === "SpedometerLow" ||
    name === "SpedometerMax" ||
    name === "SpedometerMiddle"
  )
    return true;
  if (name === "AirbudsCaseMinimlistic") return true;
  return false;
}

const icons: Record<string, IconComponent> = {};

for (const [name, component] of Object.entries(SolarIcons)) {
  if (EXCLUDED.has(name)) continue;
  if (typeof component !== "function") continue;
  if (isTypoDuplicate(name)) continue;
  icons[name] = component as IconComponent;
}

icons.Layout = SolarIcons.Widget2;
icons.Briefcase = SolarIcons.Case;
icons.FolderKanban = SolarIcons.Widget5;
icons.BookOpen = SolarIcons.Book2;
icons.Boxes = SolarIcons.Widget3;
icons.Building2 = SolarIcons.Buildings2;
icons.CircuitBoard = SolarIcons.Cpu;
icons.Cog = SolarIcons.SettingsMinimalistic;
icons.FlaskConical = SolarIcons.TestTube;
icons.GitBranch = SolarIcons.BranchingPathsUp;
icons.Globe = SolarIcons.Global;
icons.Image = SolarIcons.Gallery;
icons.Link2 = SolarIcons.Link;
icons.MessageSquare = SolarIcons.ChatSquare;
icons.Music = SolarIcons.MusicNote;
icons.Package = SolarIcons.BoxMinimalistic;
icons.PenTool = SolarIcons.Pen;
icons.Puzzle = SolarIcons.Widget6;
icons.ScrollText = SolarIcons.DocumentText;
icons.Search = SolarIcons.Magnifier;
icons.Share2 = SolarIcons.Share;
icons.ShoppingBag = SolarIcons.Bag;
icons.Smile = SolarIcons.SmileCircle;
icons.Table = SolarIcons.Widget4;
icons.Tags = SolarIcons.Tag;
icons.Terminal = SolarIcons.Programming;
icons.ThumbsUp = SolarIcons.Like;
icons.Timer = SolarIcons.Stopwatch;
icons.Tool = SolarIcons.PenNewSquare;
icons.Trash2 = SolarIcons.TrashBinMinimalistic;
icons.Trophy = SolarIcons.CupFirst;
icons.Tv2 = SolarIcons.Tv;
icons.Users = SolarIcons.UsersGroupTwoRounded;
icons.Video = SolarIcons.Videocamera;
icons.Wand2 = SolarIcons.MagicStick2;
icons.Watch = SolarIcons.WatchRound;
icons.Zap = SolarIcons.Bolt;

export default icons;
