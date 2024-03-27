FROM nginx:1.25

COPY nginx.conf /etc/nginx/nginx.conf
COPY default.conf /etc/nginx/conf.d/default.conf
RUN groupadd -r -g 1001 app && useradd -r -u 1001 -g app app && \
	touch /var/run/nginx.pid && \
    chown -R app:app /var/run/nginx.pid && \
    chown -R app:app /var/cache/nginx
EXPOSE 8080
USER app
CMD ["nginx", "-g", "daemon off;"]
