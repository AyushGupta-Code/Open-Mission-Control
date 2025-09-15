import NextAuth from "next-auth"

declare module "next-auth" {
  interface Session {
    accessToken?: string
    refreshToken?: string
    idToken?: string
  }

  interface JWT {
    accessToken?: string
    refreshToken?: string
    idToken?: string
  }
}
