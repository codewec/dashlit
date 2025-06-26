FROM node:22 AS builder
WORKDIR /app

COPY . .

RUN mv .env.example .env
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN pnpm install --frozen-lockfile --force
RUN npm run build

# second stage
FROM node:22-alpine

WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json .

EXPOSE 3000
CMD ["sh","-c","node build"]
