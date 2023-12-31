import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls'
import { OBJLoader } from 'three/examples/jsm/loaders/OBJLoader';
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader';
// CAMERA
const camera: THREE.PerspectiveCamera = new THREE.PerspectiveCamera(30, window.innerWidth / window.innerHeight, 1, 1500);
camera.position.set(-35, 70, 100);
camera.lookAt(new THREE.Vector3(0, 0, 0));

// RENDERER
const renderer: THREE.WebGLRenderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setPixelRatio(window.devicePixelRatio);
renderer.setSize(window.innerWidth, window.innerHeight);
renderer.shadowMap.enabled = true;
document.body.appendChild(renderer.domElement);

// WINDOW RESIZE HANDLING
export function onWindowResize() {
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();
  renderer.setSize(window.innerWidth, window.innerHeight);
}
window.addEventListener('resize', onWindowResize);

// SCENE
const scene: THREE.Scene = new THREE.Scene()
scene.background = new THREE.Color(0xbfd1e5);

// CONTROLS
const controls = new OrbitControls(camera, renderer.domElement);

export function animate() {
  dragObject();
  renderer.render(scene, camera);
  requestAnimationFrame(animate);
}

// ambient light
let hemiLight = new THREE.AmbientLight(0xffffff, 0.20);
scene.add(hemiLight);

//Add directional light
let dirLight = new THREE.DirectionalLight(0xffffff, 1);
dirLight.position.set(-30, 50, -30);
scene.add(dirLight);
dirLight.castShadow = true;
dirLight.shadow.mapSize.width = 2048;
dirLight.shadow.mapSize.height = 2048;
dirLight.shadow.camera.left = -70;
dirLight.shadow.camera.right = 70;
dirLight.shadow.camera.top = 70;
dirLight.shadow.camera.bottom = -70;

function createFloor() {
  // Set the position and scale of the floor
  let pos = { x: 0, y: -1, z: 3 };
  let scale = { x: 100, y: 2, z: 100 };

  // Create a box geometry for the floor
  let floorGeometry = new THREE.BoxBufferGeometry();

  // Load a texture for the floor
  let textureLoader = new THREE.TextureLoader();
  let texture = textureLoader.load('Maps.jpg'); // Replace 'path/to/your/image.jpg' with the actual path to your image file

  // Create a material with the texture
  let floorMaterial = new THREE.MeshBasicMaterial({ map: texture });

  // Create a mesh for the floor using the geometry and material
  let floorMesh = new THREE.Mesh(floorGeometry, floorMaterial);

  // Set the position and scale of the floor mesh
  floorMesh.position.set(pos.x, pos.y, pos.z);
  floorMesh.scale.set(scale.x, scale.y, scale.z);

  // Add the floor mesh to the scene
  scene.add(floorMesh);

  // Add user data to the floor mesh
  floorMesh.userData.ground = true;
}




const raycaster = new THREE.Raycaster(); // create once
const clickMouse = new THREE.Vector2();  // create once
const moveMouse = new THREE.Vector2();   // create once
var draggable: THREE.Object3D;

function intersect(pos: THREE.Vector2) {
  raycaster.setFromCamera(pos, camera);
  return raycaster.intersectObjects(scene.children);
}

window.addEventListener('click', event => {
  if (draggable != null) {
    console.log(`dropping draggable ${draggable.userData.name}`)
    draggable = null as any
    return;
  }

  // THREE RAYCASTER
  clickMouse.x = (event.clientX / window.innerWidth) * 2 - 1;
  clickMouse.y = -(event.clientY / window.innerHeight) * 2 + 1;

  const found = intersect(clickMouse);
  if (found.length > 0) {
    if (found[0].object.userData.draggable) {
      draggable = found[0].object
      console.log(`found draggable ${draggable.userData.name}`)
    }
  }
})

window.addEventListener('mousemove', event => {
  moveMouse.x = (event.clientX / window.innerWidth) * 2 - 1;
  moveMouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
});

function dragObject() {
  if (draggable != null) {
    const found = intersect(moveMouse);
    if (found.length > 0) {
      for (let i = 0; i < found.length; i++) {
        if (!found[i].object.userData.ground)
          continue
        
        let target = found[i].point;
        draggable.position.x = target.x
        draggable.position.z = target.z
      }
    }
  }
}


function loadGLTFModel() {
  const gltfLoader = new GLTFLoader();

  // Charger le modèle glTF (remplacez 'path/to/your/model.gltf' par le chemin réel de votre modèle)
  gltfLoader.load('./marrakech-tower/scene.gltf', (gltf) => {
    // Récupérer la scène du modèle glTF (contient tous les objets du modèle)
    const modelScene = gltf.scene;

    // Positionner et ajuster la taille du modèle
    modelScene.scale.set(3, 3, 3);

    // Assuming the floor is at y = -1 and has a height of 2 (adjust these values based on your floor)
    const floorY = 14;
    const floorHeight = 2;
    modelScene.castShadow = true
    modelScene.receiveShadow = true

    modelScene.userData.draggable = true
    modelScene.userData.name = 'Mosque'
    // Calculate the position to place the model on the floor
    const modelY = floorY + floorHeight / 2; // Center the model vertically on the floor

    modelScene.position.set(-15, modelY,-11);

    // Ajouter le modèle à la scène
    scene.add(modelScene);
  });
}

// Appeler la fonction pour charger le modèle glTF
loadGLTFModel();




createFloor()
animate()
