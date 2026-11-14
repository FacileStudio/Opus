<script lang="ts">
	import type { CursorState } from '@facile/mise-en-scene';

	let { cursor, color, name }: { cursor: CursorState; color: string; name: string } = $props();
</script>

<!--
@component
A collaborator's live cursor, drawn the way the board overlay draws a peer: an arrow and a name pill
filled with their colour, the name in black. Move it with `act.move` like the kit's cursor; it travels
on the same curve.
-->

<span
	class="live"
	style:translate="{cursor.x}px {cursor.y}px"
	style:opacity={cursor.visible ? 1 : 0}
	style:scale={cursor.down ? 0.88 : 1}
	style:--who={color}
>
	<svg width="16" height="17" viewBox="0 0 16 17" aria-hidden="true">
		<path d="M0 0L4 16L8 10L15 9Z" fill="var(--who)" />
	</svg>
	<span class="tag rounded-fc-pill text-fc-xs font-medium">{name}</span>
</span>

<style>
	.live {
		position: absolute;
		top: 0;
		left: 0;
		z-index: 25;
		pointer-events: none;
		transform-origin: 0 0;
		filter: drop-shadow(0 1px 2px color-mix(in oklch, var(--color-fc-scrim) 30%, transparent));
		transition:
			translate 0.9s cubic-bezier(0.45, 0, 0.1, 1),
			opacity 0.3s ease,
			scale 0.12s ease;
	}
	svg {
		display: block;
	}
	.tag {
		position: absolute;
		top: 16px;
		left: 12px;
		height: 20px;
		padding: 0 6px;
		line-height: 20px;
		white-space: nowrap;
		color: black;
		background: var(--who);
	}
	@media (prefers-reduced-motion: reduce) {
		.live {
			transition: none;
		}
	}
</style>
