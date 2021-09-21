FROM ubuntu:latest

WORKDIR /odin

RUN wget https://repo.continuum.io/miniconda/Miniconda3-latest-Linux-x86_64.sh
ENV HUGO_ID=hugo${HUGO_TYPE}_${HUGO_VERSION}
RUN wget https://github.com`wget -qO- https://github.com/gohugoio/hugo/releases/latest | grep -oE -m 1 '\/gohugoio\/hugo\/releases\/download\/v[0-9]+.[0-9]+.[0-9]*\/hugo_[0-9]+.[0-9]+.[0-9]*_Linux-64bit.deb'` | tar -xz -C /tmp \
    && mkdir -p /usr/local/sbin \
    && mv /tmp/hugo /usr/local/sbin/hugo \
    && rm -rf /tmp/${HUGO_ID}_linux_amd64 \
    && rm -rf /tmp/LICENSE.md \
    && rm -rf /tmp/README.md

# install in batch (silent) mode, does not edit PATH or .bashrc or .bash_profile
# -p path
# -f force
RUN bash Miniconda3-latest-Linux-x86_64.sh -b
    && rm Miniconda3-latest-Linux-x86_64.sh

ENV PATH=/root/miniconda3/bin:${PATH}

#RUN source /root/.bashrc
#RUN source /root/.bash_profile

RUN conda update -y conda

COPY ./requirements ./requirements

RUN conda env create --file requirements/conda.yml
    && conda clean --all --yes --quiet

COPY . .

# start the jupyter notebook in server mode
CMD python scripts/
