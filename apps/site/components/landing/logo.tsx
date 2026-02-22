export function Logo() {
  return (
    <span className="inline-flex items-center">
      <img src="/logo-dark.svg" alt="Kaneo" className="h-6 w-auto dark:hidden" />
      <img src="/logo-light.svg" alt="Kaneo" className="hidden h-6 w-auto dark:block" />
    </span>
  );
}
