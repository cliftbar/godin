mapboxgl.accessToken = 'pk.eyJ1IjoiY2xpZnRiYXIiLCJhIjoiY2t2YTk2cXIyOTB6czJ3dDl0cDJleWd3aiJ9.WC0BGBbqYty6GtxqglyUfw';
const latlngStart = [-90.19, 29.11]
const map = new mapboxgl.Map({
    container: 'map',
    style: 'mapbox://styles/mapbox/streets-v11',
    center: latlngStart,
    zoom: 6
});

map.on('render', () => {
    console.log(map.getZoom())
})

let loaded = false
map.on('load', () => {
    map.addSource('storm-data-source', {
        type: 'geojson',
        data: {
            'type': 'FeatureCollection',
            'features': []
        }
    });
    const none = ['<', ['get', 'zValue'], 34];
    const td = ['all', ['>=', ['get', 'zValue'], 34], ['<', ['get', 'zValue'], 64]];
    const cat1 = ['all', ['>=', ['get', 'zValue'], 64], ['<', ['get', 'zValue'], 83]];
    const cat2 = ['all', ['>=', ['get', 'zValue'], 83], ['<', ['get', 'zValue'], 96]];
    const cat3 = ['all', ['>=', ['get', 'zValue'], 96], ['<', ['get', 'zValue'], 113]];
    const cat4 = ['all', ['>=', ['get', 'zValue'], 113], ['<', ['get', 'zValue'], 137]];
    const cat5 = ['>=', ['get', 'zValue'], 137];

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
    loaded = true
});

function calculateStorm() {
    if (!loaded) {
        return
    }
    let maxWindKts = parseFloat(document.getElementById("inpt_maxWindKts").value) * 1.15;
    let rMaxNmi = parseFloat(document.getElementById("inpt_rMaxNmi").value) * 1.15;
    let fSpeedKts = parseFloat(document.getElementById("inpt_fSpeedKts").value) * 1.15;
    let headingDeg = parseFloat(document.getElementById("inpt_headingDeg").value) * 1.15;

    let geoJsonString = calculateLandfall(
        latlngStart[1] + 5, latlngStart[1] - 5, latlngStart[0] - 5, latlngStart[0] + 5,
        latlngStart[1], latlngStart[0],
        maxWindKts, rMaxNmi, fSpeedKts, headingDeg, 0.9,
        10, 350
    )
    console.log("out of go")
    map.getSource('storm-data-source').setData(JSON.parse(geoJsonString));

}

function updatePointSize(){
    let pointSize = parseFloat(document.getElementById("inpt_pointSize").value);
    map.setPaintProperty('storm-data-layer', 'circle-radius', pointSize)
}

function updatePointOpacity(){
    let pointOpacity = parseFloat(document.getElementById("inpt_pointOpacity").value) / 100.0;
    map.setPaintProperty('storm-data-layer', 'circle-opacity', pointOpacity)
}
