FROM scratch
EXPOSE 8080

WORKDIR /server
COPY static /server/static
COPY main /server/tas-apply

ENTRYPOINT ["./tas-apply"]
CMD [""]