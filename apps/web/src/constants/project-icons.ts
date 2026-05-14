import * as SolarIcons from "@solar-icons/react";
import type { ComponentType, SVGProps } from "react";

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>;

const EXCLUDED = new Set(["IconBase", "SolarProvider", "useSolar", "default"]);

function isTypoDuplicate(name: string): boolean {
  if (name.includes("Minimlistic")) return true;
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

function isIconComponent(value: unknown): value is IconComponent {
  if (typeof value === "function") return true;
  if (
    typeof value === "object" &&
    value !== null &&
    "$$typeof" in value &&
    "render" in value
  )
    return true;
  return false;
}

const icons: Record<string, IconComponent> = {};

for (const [name, component] of Object.entries(SolarIcons)) {
  if (EXCLUDED.has(name)) continue;
  if (!isIconComponent(component)) continue;
  if (isTypoDuplicate(name)) continue;
  icons[name] = component;
}

icons.Layout = SolarIcons.Widget2;

export default icons;
