# Use Nginx to serve the pre-built dist folder
FROM nginx:alpine

# Copy pre-built files
COPY dist /usr/share/nginx/html

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
