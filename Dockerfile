FROM scratch

COPY ./kmain /kmain

ENTRYPOINT [ "/kmain" ] 
