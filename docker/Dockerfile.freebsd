FROM auchida/freebsd

ADD bin/smartagent /usr/bin
ADD conf/client.conf /etc

RUN sed -i 's|^server.*$|server=192.168.3.147:13080|g' /etc/client.conf

CMD /usr/bin/smartagent -conf /etc/client.conf