mapboxgl.accessToken = 'pk.eyJ1IjoiY2xpZnRiYXIiLCJhIjoiY2t2YTk2cXIyOTB6czJ3dDl0cDJleWd3aiJ9.WC0BGBbqYty6GtxqglyUfw';
const lngLatStart = [-90.19, 29.11];
const emptyGeoJSON = {
    'type': 'FeatureCollection',
    'features': []
};

const map = new mapboxgl.Map({
    container: 'map',
    style: 'mapbox://styles/mapbox/streets-v11',
    center: lngLatStart,
    zoom: 6
});

// map.on('render', () => {
//     console.log(map.getZoom())
// })

let loaded = false
map.on('load', () => {
    map['doubleClickZoom'].disable()

    map.addSource('storm-data-source', {
        type: 'geojson',
        data: emptyGeoJSON
    });
    map.addSource('track-data-source', {
        type: 'geojson',
        data: emptyGeoJSON
    });

    const none = ['<', ['get', 'zValue'], 34];
    const td = ['all', ['>=', ['get', 'zValue'], 34], ['<', ['get', 'zValue'], 64]];
    const cat1 = ['all', ['>=', ['get', 'zValue'], 64], ['<', ['get', 'zValue'], 83]];
    const cat2 = ['all', ['>=', ['get', 'zValue'], 83], ['<', ['get', 'zValue'], 96]];
    const cat3 = ['all', ['>=', ['get', 'zValue'], 96], ['<', ['get', 'zValue'], 113]];
    const cat4 = ['all', ['>=', ['get', 'zValue'], 113], ['<', ['get', 'zValue'], 137]];
    // const cat5 = ['>=', ['get', 'zValue'], 137];

    map.addLayer({
        'id': 'storm-data-layer',
        'type': 'circle',
        'paint': {
            'circle-radius': 6,
            'circle-color': [
                'case',
                none,
                'gray',
                td,
                'darkgreen',
                cat1,
                'lightgreen',
                cat2,
                'yellow',
                cat3,
                'orange',
                cat4,
                'red',
                'darkred'
            ],
            'circle-opacity': 0.5,
        },
        'source': 'storm-data-source', // reference the data source
    });
    map.addLayer({
        'id': 'track-data-layer',
        'type': 'circle',
        'paint': {
            'circle-color': 'black',
            'circle-radius': 6
        },
        'source': 'track-data-source', // reference the data source
    });
    loaded = true
});

// Calculation Functions
function calculateStorm() {
    if (!loaded) {
        return;
    }
    let maxWindKts = parseFloat(document.getElementById("inpt_maxWindKts").value) / 1.15;
    let rMaxNmi = parseFloat(document.getElementById("inpt_rMaxNmi").value) / 1.15;
    let fSpeedKts = parseFloat(document.getElementById("inpt_fSpeedKts").value) / 1.15;

    let headingDeg = parseFloat(document.getElementById("inpt_headingDeg").value);

    let landfallLat = parseFloat(document.getElementById("inpt_landfallLat").value);
    let landfallLng = parseFloat(document.getElementById("inpt_landfallLng").value);
    let bboxOffset = parseFloat(document.getElementById("inpt_bboxOffset").value);

    console.time('Go calculateLandfall');
    let geoJsonString = calculateLandfall(
        landfallLat + bboxOffset, landfallLat - bboxOffset, landfallLng - bboxOffset, landfallLng + bboxOffset,
        landfallLat, landfallLng,
        maxWindKts, rMaxNmi, fSpeedKts, headingDeg, 0.9,
        10, 350
    );
    console.timeEnd('Go calculateLandfall');
    map.getSource('storm-data-source').setData(JSON.parse(geoJsonString));

}

function extrapolateTrack() {
    let landfallLat = parseFloat(document.getElementById("inpt_landfallLat").value);
    let landfallLng = parseFloat(document.getElementById("inpt_landfallLng").value);

    let headingDeg = parseFloat(document.getElementById("inpt_headingDeg").value);
    let headingXY = (90 - headingDeg) % 360
    // if (headingXY < -180) {
    //     headingXY = 360 + headingXY
    // }
    headingXY = headingXY * (Math.PI / 180)

    let degPerHour = parseFloat(document.getElementById("inpt_fSpeedKts").value) / 1.15 / 60;

    let front = [];
    for (let i = 0; i <= 24; i++) {
        let newLat = landfallLat + (degPerHour * i * Math.sin(headingXY));
        let newLon = landfallLng + (degPerHour * i * Math.cos(headingXY));
        front.push([newLat, newLon])
    }

    let back = []
    for (let i = 1; i <= 12; i++) {
        let newLat = landfallLat - (degPerHour * i * Math.sin(headingXY));
        let newLon = landfallLng - (degPerHour * i * Math.cos(headingXY));
        back.push([newLat, newLon])
    }
    let all = front.concat(back)

    let allGeoJSON = []

    all.forEach((e) => {
        allGeoJSON.push({
            "type": "Feature",
            "geometry": {
                "type": "Point",
                "coordinates": [e[1], e[0]]
            }
        })
    })
    // console.log(allGeoJSON)
    let trackFeature = {
        "type": "FeatureCollection",
        "features": allGeoJSON
    }

    map.getSource('track-data-source').setData(trackFeature);
}

// Layer Controls
map.on("dblclick", (e) => {
    document.getElementById("inpt_landfallLng").value = Math.round((e.lngLat.lng + Number.EPSILON) * 100) / 100
    document.getElementById("inpt_landfallLat").value = Math.round((e.lngLat.lat + Number.EPSILON) * 100) / 100
})

function clearStorm(){
    map.getSource('storm-data-source').setData(emptyGeoJSON);
    map.getSource('track-data-source').setData(emptyGeoJSON);
}

function updatePointSize(){
    let pointSize = parseFloat(document.getElementById("inpt_pointSize").value);
    map.setPaintProperty('storm-data-layer', 'circle-radius', pointSize);
}

function updatePointOpacity(){
    let pointOpacity = parseFloat(document.getElementById("inpt_pointOpacity").value) / 100.0;
    map.setPaintProperty('storm-data-layer', 'circle-opacity', pointOpacity);
}


