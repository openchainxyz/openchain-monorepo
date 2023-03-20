# Install dependencies only when needed
FROM node:16-alpine AS deps

# Check https://github.com/nodejs/docker-node/tree/b4117f9333da4138b03a546ec926ef50a31506c3#nodealpine to understand why libc6-compat might be needed.
RUN apk add --no-cache libc6-compat && npm install -g pnpm
WORKDIR /app
COPY package.json ./
COPY pnpm-lock.yaml ./
RUN pnpm install

# Rebuild the source code only when needed
FROM node:16-alpine AS builder
WORKDIR /app
RUN apk add --no-cache libc6-compat && npm install -g pnpm
COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN pnpm run build

# Production image, copy all the files and run next
FROM node:19-alpine AS runner
WORKDIR /app

ENV NODE_ENV production

RUN apk add --no-cache tini

COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist

EXPOSE 3000

ENV PORT 3000

WORKDIR /app/dist

ENTRYPOINT ["tini", "--"]

CMD ["node", "main.js"]
