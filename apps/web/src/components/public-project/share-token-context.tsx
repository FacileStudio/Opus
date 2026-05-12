import { createContext, useContext } from "react";

const ShareTokenContext = createContext<string | null>(null);

export const ShareTokenProvider = ShareTokenContext.Provider;

export function useShareToken() {
  return useContext(ShareTokenContext);
}
