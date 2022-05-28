FROM scratch

COPY ./imain /imain

ENTRYPOINT [ "/imain" ] 
