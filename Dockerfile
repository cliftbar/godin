FROM ubuntu:latest

WORKDIR /odin

RUN wget https://repo.continuum.io/miniconda/Miniconda3-latest-Linux-x86_64.sh
ENV HUGO_VERSION='0.88.1'
ENV HUGO_NAME="hugo_extended_${HUGO_VERSION}_Linux-64bit"
ENV HUGO_URL="https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_NAME}.deb"
ENV BUILD_DEPS="wget"

RUN apt-get update && \
    apt-get install -y git "${BUILD_DEPS}" && \
    wget "${HUGO_URL}" && \
    apt-get install "./${HUGO_NAME}.deb" && \
    rm -rf "./${HUGO_NAME}.deb" "${HUGO_NAME}" && \
    apt-get remove -y "${BUILD_DEPS}" && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN bash Miniconda3-latest-Linux-x86_64.sh -b
    && rm Miniconda3-latest-Linux-x86_64.sh

ENV PATH=/root/miniconda3/bin:${PATH}

RUN conda update -y conda

COPY ./requirements ./requirements

RUN conda env create --file requirements/conda.yml
    && conda clean --all --yes --quiet

COPY . .

CMD python scripts/
