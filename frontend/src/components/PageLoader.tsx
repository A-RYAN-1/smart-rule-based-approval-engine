/** Full-screen spinner shown while the session is checked or a page chunk loads. */
export function PageLoader() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-background">
      <div role="status" className="flex flex-col items-center gap-4">
        <div aria-hidden="true" className="h-10 w-10 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        <p className="text-muted-foreground">Loading...</p>
      </div>
    </main>
  );
}
