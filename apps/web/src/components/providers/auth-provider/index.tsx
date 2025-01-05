import useGetMe from "@/hooks/queries/use-get-me";
import type { User } from "@/types/user";
import { LayoutGrid } from "lucide-react";
import {
  type PropsWithChildren,
  createContext,
  useEffect,
  useState,
} from "react";

export const AuthContext = createContext<{
  user: User | null;
  setUser: (user: User | null) => void;
}>({
  user: null,
  setUser: () => {},
});

function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<{
    name: string;
    id: string;
    password: string;
    email: string;
    createdAt: Date;
  } | null>(null);

  const { data, isFetching } = useGetMe(user?.id);

  useEffect(() => {
    if (data) {
      setUser(data.data);
    }
  }, [data]);

  if (isFetching) {
    return (
      <div className="flex w-full items-center justify-center h-screen flex-col md:flex-row bg-zinc-50 dark:bg-zinc-950">
        <div className="p-1.5 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-lg shadow-sm animate-spin">
          <LayoutGrid className="w-5 h-5 text-white" />
        </div>
      </div>
    );
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        setUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export default AuthProvider;
