const none = ['<', ['get', 'zValue'], 34];
const td = ['all', ['>=', ['get', 'zValue'], 34], ['<', ['get', 'zValue'], 64]];
const cat1 = ['all', ['>=', ['get', 'zValue'], 64], ['<', ['get', 'zValue'], 83]];
const cat2 = ['all', ['>=', ['get', 'zValue'], 83], ['<', ['get', 'zValue'], 96]];
const cat3 = ['all', ['>=', ['get', 'zValue'], 96], ['<', ['get', 'zValue'], 113]];
const cat4 = ['all', ['>=', ['get', 'zValue'], 113], ['<', ['get', 'zValue'], 137]];
// const cat5 = ['>=', ['get', 'zValue'], 137];

async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

mapboxgl.accessToken = 'pk.eyJ1IjoiY2xpZnRiYXIiLCJhIjoiY2t2ZnduM2l4MjNiYzJvdDJ0eWx2YWF2dSJ9.zKxbrzsoNUpkfA3MQ7EO6Q';
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
let allTrackCoords = []
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
    let ptrMap = calculateLandfall(
        landfallLat + bboxOffset, landfallLat - bboxOffset, landfallLng - bboxOffset, landfallLng + bboxOffset,
        landfallLat, landfallLng,
        maxWindKts, rMaxNmi, fSpeedKts, headingDeg, 0.9,
        10, 350,
        -1
    );
    console.log("out of go")
    console.timeEnd('Go calculateLandfall');



    console.time("decode")
    // let s = "";
    // for (let i = ptrMap.ptr; i < ptrMap.ptr + ptrMap.len; ++i)
    //     s += String.fromCharCode(wasmMemory[i]);
    let strDecode = new TextDecoder("utf-8").decode(result.instance.exports.mem.buffer.slice(ptrMap.ptr, ptrMap.ptr + ptrMap.len))

    console.timeEnd("decode")

    console.time('Map setData');
    map.getSource('storm-data-source').setData(JSON.parse(strDecode));
    // map.getSource('storm-data-source').setData("data:application/json,"+geoJsonString);
    console.timeEnd('Map setData');

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
    let maxWindKts = parseFloat(document.getElementById("inpt_maxWindKts").value) / 1.15;

    let front = [];
    for (let i = 0; i <= 24; i++) {
        let newLat = landfallLat + (degPerHour * i * Math.sin(headingXY));
        let newLon = landfallLng + (degPerHour * i * Math.cos(headingXY));
        let newWind = (((maxWindKts / 3 - maxWindKts) / (24 - 0)) * (i - 0)) + maxWindKts
        front.push([newLat, newLon, i, newWind])
    }

    let back = []
    for (let i = 1; i <= 12; i++) {
        let newLat = landfallLat - (degPerHour * i * Math.sin(headingXY));
        let newLon = landfallLng - (degPerHour * i * Math.cos(headingXY));

        back.push([newLat, newLon, 0-i, maxWindKts])
    }
    allTrackCoords = front.concat(back)

    let allGeoJSON = []

    allTrackCoords.forEach((elm) => {
        allGeoJSON.push({
            "type": "Feature",
            "geometry": {
                "type": "Point",
                "coordinates": [elm[1], elm[0]]
            },
            "properties": {
                "hour": elm[2]
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

let animationReq = undefined;
async function calculateStormAnimation() {
    if (!loaded || animationReq !== undefined) {
        return;
    }
    // let maxWindKts = parseFloat(document.getElementById("inpt_maxWindKts").value) / 1.15;
    let rMaxNmi = parseFloat(document.getElementById("inpt_rMaxNmi").value) / 1.15;
    let fSpeedKts = parseFloat(document.getElementById("inpt_fSpeedKts").value) / 1.15;

    let headingDeg = parseFloat(document.getElementById("inpt_headingDeg").value);

    let bboxOffset = parseFloat(document.getElementById("inpt_bboxOffset").value);

    allTrackCoords.sort((firstEl, secondEl) => { return firstEl[2] - secondEl[2] } )

    let sourceName = 'animation-data-source'
    let layerName = 'animation-data-layer'

    if (map.getSource(sourceName) === undefined) {
        map.addSource(sourceName, {
            type: 'geojson',
            data: emptyGeoJSON
        });
    }

    if (map.getLayer(layerName) === undefined) {
        map.addLayer({
                'id': layerName,
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
                'source': sourceName // reference the data source
            },
            'track-data-layer'
        );
    }
    // for (let i = 0; i < allTrackCoords.length; ++i) {
    let iter = 0;
    async function animationLoop(timestamp) {
        let coord = allTrackCoords[iter]
        // let sourceName = 'animation-data-source_' + coord[2]
        // let layerName = 'animation-data-layer_' + coord[2]

        let landfallLat = coord[0]
        let landfallLng = coord[1]
        let maxWindKts = coord[3]

        console.time('Go calculateLandfall');
        let geoJsonString = calculateLandfall(
            landfallLat + bboxOffset, landfallLat - bboxOffset, landfallLng - bboxOffset, landfallLng + bboxOffset,
            landfallLat, landfallLng,
            maxWindKts, rMaxNmi, fSpeedKts, headingDeg, 0.9,
            100, 350,
            30
        );
        console.timeEnd('Go calculateLandfall');

        console.time('Map setData');
        map.getSource(sourceName).setData(JSON.parse(geoJsonString));
        // map.getSource(sourceName).setData("data:application/json,base64,"+geoJsonString);
        console.timeEnd('Map setData');



        await sleep(500);
        iter = iter + 1;
        animationReq = window.requestAnimationFrame(animationLoop)
        if (allTrackCoords.length <= iter) {
            window.cancelAnimationFrame(animationReq);
            animationReq = undefined;
        }
    }

    await animationLoop(0);

}

// Layer Controls
map.on("dblclick", (e) => {
    document.getElementById("inpt_landfallLng").value = Math.round((e.lngLat.lng + Number.EPSILON) * 100) / 100
    document.getElementById("inpt_landfallLat").value = Math.round((e.lngLat.lat + Number.EPSILON) * 100) / 100
})

function clearStorm(){
    map.getSource('storm-data-source').setData(emptyGeoJSON);
    map.getSource('track-data-source').setData(emptyGeoJSON);

    allTrackCoords = [];
    map.getSource('animation-data-source').setData(emptyGeoJSON);
}

function updatePointSize(){
    let pointSize = parseFloat(document.getElementById("inpt_pointSize").value);
    map.setPaintProperty('storm-data-layer', 'circle-radius', pointSize);
}

function updatePointOpacity(){
    let pointOpacity = parseFloat(document.getElementById("inpt_pointOpacity").value) / 100.0;
    map.setPaintProperty('storm-data-layer', 'circle-opacity', pointOpacity);
}


