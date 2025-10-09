FROM docker.1ms.run/node:20-alpine AS builder

WORKDIR /app

ENV VITE_API_BASE_URL=

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM docker.1ms.run/nginx:1.27-alpine

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
