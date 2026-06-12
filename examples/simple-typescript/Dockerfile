FROM node:22

RUN apt-get update -y && apt-get install -y openssl

WORKDIR /app
RUN npm install -g pnpm
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY src ./src
COPY tsconfig.json ./
RUN pnpm run build

CMD ["pnpm", "start"]
