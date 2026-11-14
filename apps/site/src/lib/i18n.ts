import { type Catalogs, createTranslator, LOCALE_COOKIE, type Locale, resolveLocale } from '@facile/muse';

const en = {
	'nav.features': 'Features',
	'nav.download': 'Download',
	'hero.title': 'Project management that stays fast',
	'hero.text': 'Boards, backlog and timelines your whole team follows live.',
	'hero.cli': 'Get the CLI',
	'demo.label': 'An Opus board: a task card moves from In progress to Done.',
	'demo.project': 'Website',
	'demo.board': 'Board',
	'demo.todo': 'To do',
	'demo.doing': 'In progress',
	'demo.done': 'Done',
	'demo.task': 'Ship the landing page',
	'features.title': 'What a project needs',
	'features.views.title': 'Four views, one backlog',
	'features.views.text': 'Board, list, Gantt and backlog show the same tasks.',
	'features.live.title': 'Live for everyone',
	'features.live.text': 'Moves, comments and labels reach every open tab at once.',
	'features.rules.title': 'Workflow rules',
	'features.rules.text': 'The server applies status changes, assignments and due-date reminders.',
	'features.github.title': 'GitHub in the loop',
	'features.github.text': 'A merged pull request moves its task to Done.',
	'tiles.title': 'More than a board',
	'tiles.search.title': 'Find anything',
	'tiles.search.body': 'One search covers tasks, projects and comments.',
	'tiles.search.query': 'landing',
	'tiles.search.tasks': 'Tasks',
	'tiles.search.comments': 'Comments',
	'tiles.search.comment': 'Landing copy is ready for review',
	'tiles.timeline.title': 'Plan on a timeline',
	'tiles.timeline.body': 'Stretch a Gantt bar to move its due date.',
	'tiles.timeline.task': 'Beta launch',
	'tiles.github.title': 'Merged means done',
	'tiles.github.body': 'Opus follows the linked pull request and closes the task.',
	'tiles.github.task': 'Fix the signup form',
	'tiles.github.open': 'Open',
	'tiles.github.merged': 'Merged',
	'download.title': 'Opus in your terminal',
	'download.text': 'Add, list and close tasks without leaving the shell.',
	'downloads.title': 'Download Opus',
	'downloads.text': 'The Opus CLI for macOS and Linux.',
	'cta.title': 'Plan your next project in Opus',
	'cta.text': 'Sign in with your Facile account.',
	'footer.tagline': 'Project management for Facile teams.'
};

const fr: Record<keyof typeof en, string> = {
	'nav.features': 'Fonctionnalités',
	'nav.download': 'Télécharger',
	'hero.title': 'La gestion de projet qui reste rapide',
	'hero.text': 'Tableaux, backlog et calendriers que toute l’équipe suit en direct.',
	'hero.cli': 'Installer la CLI',
	'demo.label': 'Un tableau Opus : une tâche passe de En cours à Terminé.',
	'demo.project': 'Site web',
	'demo.board': 'Tableau',
	'demo.todo': 'À faire',
	'demo.doing': 'En cours',
	'demo.done': 'Terminé',
	'demo.task': 'Mettre en ligne la page d’accueil',
	'features.title': 'Ce qu’il faut à un projet',
	'features.views.title': 'Quatre vues, un backlog',
	'features.views.text': 'Tableau, liste, Gantt et backlog affichent les mêmes tâches.',
	'features.live.title': 'En direct pour tous',
	'features.live.text': 'Déplacements, commentaires et étiquettes arrivent aussitôt dans chaque onglet ouvert.',
	'features.rules.title': 'Règles de workflow',
	'features.rules.text': 'Le serveur applique statuts, assignations et rappels d’échéance.',
	'features.github.title': 'GitHub dans la boucle',
	'features.github.text': 'Une pull request fusionnée passe sa tâche en Terminé.',
	'tiles.title': 'Plus qu’un tableau',
	'tiles.search.title': 'Tout retrouver',
	'tiles.search.body': 'Une seule recherche couvre tâches, projets et commentaires.',
	'tiles.search.query': 'accueil',
	'tiles.search.tasks': 'Tâches',
	'tiles.search.comments': 'Commentaires',
	'tiles.search.comment': 'Le texte de l’accueil est prêt à relire',
	'tiles.timeline.title': 'Planifier sur une frise',
	'tiles.timeline.body': 'Étirez une barre du Gantt pour décaler son échéance.',
	'tiles.timeline.task': 'Lancement bêta',
	'tiles.github.title': 'Fusionnée, donc terminée',
	'tiles.github.body': 'Opus suit la pull request liée et clôt la tâche.',
	'tiles.github.task': 'Corriger le formulaire d’inscription',
	'tiles.github.open': 'Ouvert',
	'tiles.github.merged': 'Fusionné',
	'download.title': 'Opus dans votre terminal',
	'download.text': 'Ajoutez, listez et fermez des tâches sans quitter le shell.',
	'downloads.title': 'Télécharger Opus',
	'downloads.text': 'La CLI Opus pour macOS et Linux.',
	'cta.title': 'Planifiez votre prochain projet dans Opus',
	'cta.text': 'Connectez-vous avec votre compte Facile.',
	'footer.tagline': 'La gestion de projet des équipes Facile.'
};

const es: Record<keyof typeof en, string> = {
	'nav.features': 'Funciones',
	'nav.download': 'Descargar',
	'hero.title': 'Gestión de proyectos que sigue siendo rápida',
	'hero.text': 'Tableros, backlog y cronogramas que todo el equipo sigue en directo.',
	'hero.cli': 'Instalar la CLI',
	'demo.label': 'Un tablero de Opus: una tarea pasa de En curso a Hecho.',
	'demo.project': 'Sitio web',
	'demo.board': 'Tablero',
	'demo.todo': 'Por hacer',
	'demo.doing': 'En curso',
	'demo.done': 'Hecho',
	'demo.task': 'Publicar la página de inicio',
	'features.title': 'Lo que necesita un proyecto',
	'features.views.title': 'Cuatro vistas, un backlog',
	'features.views.text': 'Tablero, lista, Gantt y backlog muestran las mismas tareas.',
	'features.live.title': 'En directo para todos',
	'features.live.text': 'Movimientos, comentarios y etiquetas llegan al instante a cada pestaña abierta.',
	'features.rules.title': 'Reglas de flujo',
	'features.rules.text': 'El servidor aplica estados, asignaciones y recordatorios de vencimiento.',
	'features.github.title': 'GitHub en el circuito',
	'features.github.text': 'Una pull request fusionada pasa su tarea a Hecho.',
	'tiles.title': 'Más que un tablero',
	'tiles.search.title': 'Encuéntralo todo',
	'tiles.search.body': 'Una sola búsqueda cubre tareas, proyectos y comentarios.',
	'tiles.search.query': 'inicio',
	'tiles.search.tasks': 'Tareas',
	'tiles.search.comments': 'Comentarios',
	'tiles.search.comment': 'El texto de inicio está listo para revisar',
	'tiles.timeline.title': 'Planifica en un cronograma',
	'tiles.timeline.body': 'Estira una barra del Gantt para mover su fecha límite.',
	'tiles.timeline.task': 'Lanzamiento beta',
	'tiles.github.title': 'Mergeada, tarea hecha',
	'tiles.github.body': 'Opus sigue la pull request vinculada y cierra la tarea.',
	'tiles.github.task': 'Arreglar el formulario de registro',
	'tiles.github.open': 'Abierta',
	'tiles.github.merged': 'Mergeada',
	'download.title': 'Opus en tu terminal',
	'download.text': 'Crea, lista y cierra tareas sin salir de la shell.',
	'downloads.title': 'Descargar Opus',
	'downloads.text': 'La CLI de Opus para macOS y Linux.',
	'cta.title': 'Planifica tu próximo proyecto en Opus',
	'cta.text': 'Inicia sesión con tu cuenta de Facile.',
	'footer.tagline': 'Gestión de proyectos para equipos Facile.'
};

const de: Record<keyof typeof en, string> = {
	'nav.features': 'Funktionen',
	'nav.download': 'Download',
	'hero.title': 'Projektmanagement, das schnell bleibt',
	'hero.text': 'Boards, Backlog und Zeitpläne, die das ganze Team live verfolgt.',
	'hero.cli': 'CLI installieren',
	'demo.label': 'Ein Opus-Board: eine Aufgabe wandert von In Arbeit nach Erledigt.',
	'demo.project': 'Website',
	'demo.board': 'Board',
	'demo.todo': 'Offen',
	'demo.doing': 'In Arbeit',
	'demo.done': 'Erledigt',
	'demo.task': 'Landingpage veröffentlichen',
	'features.title': 'Was ein Projekt braucht',
	'features.views.title': 'Vier Ansichten, ein Backlog',
	'features.views.text': 'Board, Liste, Gantt und Backlog zeigen dieselben Aufgaben.',
	'features.live.title': 'Live für alle',
	'features.live.text': 'Verschobene Aufgaben, Kommentare und Labels erscheinen sofort in jedem offenen Tab.',
	'features.rules.title': 'Workflow-Regeln',
	'features.rules.text': 'Der Server setzt Status, Zuweisungen und Fristerinnerungen um.',
	'features.github.title': 'GitHub eingebunden',
	'features.github.text': 'Ein gemergter Pull Request setzt seine Aufgabe auf Erledigt.',
	'tiles.title': 'Mehr als ein Board',
	'tiles.search.title': 'Alles finden',
	'tiles.search.body': 'Eine Suche findet Aufgaben, Projekte und Kommentare.',
	'tiles.search.query': 'landing',
	'tiles.search.tasks': 'Aufgaben',
	'tiles.search.comments': 'Kommentare',
	'tiles.search.comment': 'Der Landing-Text ist bereit zur Prüfung',
	'tiles.timeline.title': 'Auf der Zeitachse planen',
	'tiles.timeline.body': 'Ziehen Sie einen Gantt-Balken, um die Frist zu verschieben.',
	'tiles.timeline.task': 'Beta-Start',
	'tiles.github.title': 'Gemergt heißt erledigt',
	'tiles.github.body': 'Opus verfolgt den verknüpften Pull Request und schließt die Aufgabe.',
	'tiles.github.task': 'Anmeldeformular reparieren',
	'tiles.github.open': 'Offen',
	'tiles.github.merged': 'Zusammengeführt',
	'download.title': 'Opus im Terminal',
	'download.text': 'Aufgaben anlegen, auflisten und schließen, ohne die Shell zu verlassen.',
	'downloads.title': 'Opus herunterladen',
	'downloads.text': 'Die Opus-CLI für macOS und Linux.',
	'cta.title': 'Planen Sie Ihr nächstes Projekt in Opus',
	'cta.text': 'Melden Sie sich mit Ihrem Facile-Konto an.',
	'footer.tagline': 'Projektmanagement für Facile-Teams.'
};

const catalogs: Catalogs<typeof en> = { en, fr, es, de };

let active: () => Locale = () => 'en';

/** The picked locale cookie, then the browser language. */
export function localeSource(): string | null {
	const prefix = `${LOCALE_COOKIE}=`;
	const found = document.cookie.split('; ').find((part) => part.startsWith(prefix));
	return found ? decodeURIComponent(found.slice(prefix.length)) : navigator.language;
}

/** Keeps a locale picked in the footer for the next visit. */
export function rememberLocale(locale: Locale) {
	document.cookie = `${LOCALE_COOKIE}=${locale}; path=/; max-age=31536000; samesite=lax`;
}

/** Points the translator at the locale `provideLocale` resolved. */
export function bindLocale(source: () => Locale) {
	active = () => resolveLocale(source());
}

/** Translates a site message in the current locale. */
export const t = createTranslator(catalogs, () => active());
