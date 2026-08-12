export type User = {
  id: string;
  name: string;
  email: string;
  emailVerified: boolean;
  image: string;
  color: string;
  locale: string;
  createdAt: string;
  updatedAt: string;
};

export type Session = {
  user: User;
};