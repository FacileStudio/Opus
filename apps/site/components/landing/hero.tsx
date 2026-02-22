"use client";

import { Github } from "lucide-react";
import { Button } from "@/components/ui/button";

export function Hero() {
  const openPreview = () => {
    const isDark = document.documentElement.classList.contains("dark");
    const previewSrc = isDark ? "/images/dark.png" : "/images/light.png";
    window.open(previewSrc, "_blank", "noopener,noreferrer");
  };

  return (
    <section className="relative overflow-hidden px-6 pt-14 pb-16 md:pt-20 md:pb-20 lg:pt-24">
      <div className="mx-auto grid w-full max-w-7xl gap-8 lg:grid-cols-[0.72fr_1.28fr] lg:items-center">
        <div className="lg:max-w-md">
          <h1 className="max-w-xl text-balance text-4xl font-medium leading-[1.06] md:text-5xl lg:text-6xl">
            All you <span className="text-primary">need</span>. Nothing you
            don&apos;t.
          </h1>
          <p className="mb-8 mt-5 max-w-xl text-balance text-lg text-muted-foreground leading-relaxed md:text-xl">
            Kaneo gives you clean planning, focused execution, and full
            ownership of your workflow from backlog to release.
          </p>

          <div className="flex flex-wrap items-center gap-3">
            <Button
              size="lg"
              className="gap-2"
              onClick={() => {
                window.location.href = "https://cloud.kaneo.app";
              }}
            >
              Cloud
            </Button>
            <Button
              variant="outline"
              size="lg"
              className="gap-2"
              onClick={() => {
                window.location.href = "/docs/core";
              }}
            >
              Get Started
            </Button>
            <Button
              variant="outline"
              size="lg"
              className="gap-2"
              onClick={() => {
                window.location.href = "https://github.com/usekaneo/kaneo";
              }}
            >
              <Github className="h-4 w-4" />
              GitHub
            </Button>
          </div>
        </div>

        <div className="relative">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-0 -z-10 translate-x-4 translate-y-4 rounded-2xl border border-border/50 bg-muted/20"
          />
          <button
            type="button"
            onClick={openPreview}
            className="group relative w-full overflow-hidden rounded-2xl border border-border/70 bg-card p-2 text-left shadow-[0_30px_80px_-44px_rgba(0,0,0,0.55)] transition duration-300 ease-out"
            aria-label="Open full-size preview"
          >
            <img
              src="/images/dark.png"
              alt="Kaneo app preview"
              className="hidden h-[560px] w-full rounded-xl border border-border/60 object-cover object-left-top dark:block"
            />
            <img
              src="/images/light.png"
              alt="Kaneo app preview"
              className="block h-[560px] w-full rounded-xl border border-border/60 object-cover object-left-top dark:hidden"
            />
          </button>
        </div>
      </div>
    </section>
  );
}
