mapboxgl.accessToken = 'pk.eyJ1IjoiY2xpZnRiYXIiLCJhIjoiY2t2YTk2cXIyOTB6czJ3dDl0cDJleWd3aiJ9.WC0BGBbqYty6GtxqglyUfw';
const map = new mapboxgl.Map({
    container: 'map',
    style: 'mapbox://styles/mapbox/streets-v11',
    center: [-74.5, 40],
    zoom: 9
});

// const canvas = map.getCanvas() // document.getElementById('canvasID');
// const ctx = canvas.getContext('2d');
let canvas; // document.getElementById('canvasID');
let ctx;
const circles = [];
// if (ctx) {
//     canvas.style.display = 'none';
// }

// map.on('load', () => {
//     map.addLayer({
//         id: 'terrain-data',
//         type: 'line',
//         source: {
//             type: 'vector',
//             url: 'mapbox://mapbox.mapbox-terrain-v2'
//         },
//         'source-layer': 'contour'
//     });
// });

// 39°57'49.5"N 74°04'03.7"W
// map.on('load', () => {
//
// });
let loaded = false
map.on('render', () => {
if (!loaded) {
        return
    }
    let coordinates = [
        map.getBounds().getNorthWest().toArray(),
        map.getBounds().getNorthEast().toArray(),
        map.getBounds().getSouthEast().toArray(),
        map.getBounds().getSouthWest().toArray()
    ]
    map.getSource('canvas-source').setCoordinates(coordinates);

    ctx.clearRect(0,0,canvas.width,canvas.height);
    for (let r = 0; r < circles.length; r++) {
        circles[r].update();
    }
    map.getSource('canvas-source').play()
});

map.on('load', () => {
    canvas = map.getCanvasContainer(); // document.getElementById('canvasID');
    ctx = canvas.getContext('2d');

    map.addSource('canvas-source', {
        type: 'canvas',
        canvas: 'canvasID',
        coordinates: [
            map.getBounds().getNorthWest().toArray(),
            map.getBounds().getNorthEast().toArray(),
            map.getBounds().getSouthEast().toArray(),
            map.getBounds().getSouthWest().toArray()
        ],
// Set to true if the canvas source is animated. If the canvas is static, animate should be set to false to improve performance.
        animate: false
    });

    map.addLayer({
        id: 'canvas-layer',
        type: 'raster',
        source: 'canvas-source'
    });
    loaded = true
});

function load() {
    let coords = calculateLandfall(45, 35, -80, -70,  50, -74,  100, 15, 5, 315, 0.9,  10, 350)
    coords.forEach(p => {
        circles.push(new Circle(p[0], p[1], p[2], 4))
    })
    // for (const p in coords) {
    //     circles.push(new Circle(p[0], p[1], p[2], 6))
    // }
    // map.addSource('storm-data-source', {
    //     type: 'geojson',
    //     data: {
    //         "type": "MultiPoint",
    //         "coordinates": coords
    //     }
    // });
    // map.addLayer({
    //     'id': 'storm-data-layer',
    //     'type': 'circle',
    //     'paint': {
    //         'circle-radius': 6,
    //         'circle-color': '#B42222'
    //     },
    //     // 'layout': {
    //     //     'text-field': ['get', 'coordinates']
    //     // },
    //     'source': 'storm-data-source', // reference the data source
    // });
}

function Circle(x, y, z, radius) {
    if (0 < z ){
        console.log("Excite on z " + x + "at : " + x + ", " + y)
    }
    this.x = x;
    this.y = y;

    this.radius = radius;

    let color = 'black'
    if (z === 0) {
        color = 'gray'
    }
    else if (z < 20) {
        color = 'green'
    } else if (z < 40) {
        color = 'cyan'
    } else if (z < 60) {
        color = 'yellow'
    } else if (z < 80) {
        color = 'orange'
    } else {
        color = 'red'
    }

    this.draw = function () {
        let canvasPoint = map.project([this.x, this.y])
        ctx.beginPath();
        ctx.arc(canvasPoint.x, canvasPoint.y, this.radius, 0, Math.PI * 2, false);
        ctx.strokeStyle = color;
        ctx.stroke();
    };

    this.update = function () {
        // if (this.x + this.radius > 400 || this.x - this.radius < 0) {
        //     this.dx = -this.dx;
        // }
        //
        // if (this.y + this.radius > 400 || this.y - this.radius < 0) {
        //     this.dy = -this.dy;
        // }
        //
        // this.x += this.dx;
        // this.y += this.dy;

        this.draw();
    };
}
