// app/api/auth/[...nextauth]/route.ts
import NextAuth from "next-auth"
import KeycloakProvider from "next-auth/providers/keycloak"

// 🔄 Helper to refresh access tokens with Keycloak
async function refreshAccessToken(token: any) {
  try {
    const res = await fetch(
      "http://localhost:8081/realms/open-mission-control/protocol/openid-connect/token",
      {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({
          client_id: "open-mission-control-frontend",
          client_secret: process.env.KEYCLOAK_CLIENT_SECRET!, // keep secret in env
          grant_type: "refresh_token",
          refresh_token: token.refreshToken,
        }),
      }
    )

    const refreshed = await res.json()

    if (!res.ok) {
      throw refreshed
    }

    return {
      ...token,
      accessToken: refreshed.access_token,
      accessTokenExpires: Date.now() + refreshed.expires_in * 1000, // ⏳ new expiry
      refreshToken: refreshed.refresh_token ?? token.refreshToken, // fallback
    }
  } catch (error) {
    console.error("Error refreshing access token", error)
    return {
      ...token,
      error: "RefreshAccessTokenError",
    }
  }
}

const handler = NextAuth({
  providers: [
    KeycloakProvider({
      clientId: "open-mission-control-frontend",
      clientSecret: process.env.KEYCLOAK_CLIENT_SECRET!,
      issuer: "http://localhost:8081/realms/open-mission-control",
    }),
  ],
  callbacks: {
    // 1️⃣ Handle JWT lifecycle
    async jwt({ token, user, account }) {
      if (account && user) {
        return {
          accessToken: account.access_token,
          accessTokenExpires: Date.now() + (account.expires_in as number) * 1000,
          refreshToken: account.refresh_token,
          user,
        }
      }

      if (Date.now() < (token.accessTokenExpires as number)) {
        return token
      }

      return await refreshAccessToken(token)
    },

    // 2️⃣ Expose session to frontend
    async session({ session, token }) {
      session.user = token.user as any
      session.accessToken = token.accessToken as string
      session.refreshToken = token.refreshToken as string
      session.error = token.error
      return session
    },
  },
  secret: process.env.NEXTAUTH_SECRET,
})

export { handler as GET, handler as POST }
