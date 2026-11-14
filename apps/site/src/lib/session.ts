const SSO_ATTEMPT_KEY = 'opus.sso-attempt';
const ATTEMPT_WINDOW_MS = 5 * 60 * 1000;

function ssoAttemptPending(now = Date.now()): boolean {
	try {
		const last = Number(sessionStorage.getItem(SSO_ATTEMPT_KEY));
		return Boolean(last) && now - last < ATTEMPT_WINDOW_MS;
	} catch {
		return false;
	}
}

/**
 * Where a visitor of the public pages belongs instead, or null to stay. A live
 * session goes to the app, and a tab whose silent SSO attempt just came back
 * signed out goes to the sign-in page, as the app's own index route does.
 */
export async function elsewhere(): Promise<string | null> {
	try {
		const response = await fetch('/api/auth/me', {
			credentials: 'same-origin'
		});
		if (response.ok) return '/dashboard';
	} catch {
		return null;
	}
	return ssoAttemptPending() ? '/auth/sign-in?ended=true' : null;
}
