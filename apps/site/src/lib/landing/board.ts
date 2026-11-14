import type { Act, CursorState, Point } from '@facile/mise-en-scene';
import { USER_COLORS } from '@facile/muse';
import { tick } from 'svelte';

export const LEA_COLOR = USER_COLORS[1];

export type Phase = 'idle' | 'lifted' | 'dragging' | 'done';

/** One card's trip from its column to another, driven by one person's cursor. */
export interface Move {
	phase: Phase;
	offset: Point;
	height: number;
}

/** A card still in its first column. */
export function idleMove(height: number): Move {
	return { phase: 'idle', offset: { x: 0, y: 0 }, height };
}

/** Brings a hidden cursor in from below the card and onto its grip point. */
export async function reach(act: Act, who: CursorState, card: HTMLElement | undefined): Promise<Point> {
	const grip = act.point(card, 0.35, 0.45);
	await act.move(who, { x: grip.x + 70, y: grip.y + 90 }, 400);
	who.visible = true;
	await act.move(who, grip);
	return grip;
}

/** Lifts the card at `grip`, opens a slot in the target column and drops the card into it. */
export async function drag(
	act: Act,
	who: CursorState,
	move: Move,
	grip: Point,
	card: () => HTMLElement | undefined,
	slot: () => HTMLElement | undefined
) {
	who.down = true;
	move.height = card()?.offsetHeight ?? move.height;
	move.phase = 'lifted';
	await tick();
	await act.wait(250);
	const from = act.point(card(), 0, 0);
	const to = act.point(slot(), 0, 0);
	move.offset = { x: to.x - from.x, y: to.y - from.y };
	move.phase = 'dragging';
	await act.move(who, { x: grip.x + move.offset.x, y: grip.y + move.offset.y });
	who.down = false;
	move.phase = 'done';
}

/** Moves a cursor off what it just dropped and hides it. */
export async function leave(act: Act, who: CursorState) {
	await act.move(who, { x: who.x + 60, y: who.y + 110 }, 600);
	who.visible = false;
}
