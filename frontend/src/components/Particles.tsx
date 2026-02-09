import { useEffect, useRef } from 'react';
import * as THREE from 'three/webgpu';
import {
  Fn,
  uniform,
  storage,
  instanceIndex,
  float,
  vec3,
  vec4,
  If,
  Loop,
  color,
  smoothstep,
  mix,
  cross,
  mod,
  normalize,
  length,
  sub,
  mul,
  add,
  div,
  max,
} from 'three/tsl';

export function Particles() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let disposed = false;
    let initialized = false;
    let isVisible = true;

    // -- State --
    const mouse = new THREE.Vector2(0, 0);
    let mouseOnCanvas = false;
    let mouseAttractorStrength = 0;
    const mouseWorld = new THREE.Vector3(0, 0, 0);
    let currentScrollY = window.scrollY;
    let currentCameraAngle = 0;

    // -- Scene --
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(60, 1, 0.1, 100);
    camera.position.set(0, 3, 5);
    camera.lookAt(0, 0, 0);

    const ambientLight = new THREE.AmbientLight('#ffffff', 0.5);
    scene.add(ambientLight);
    const directionalLight = new THREE.DirectionalLight('#ffffff', 1.5);
    directionalLight.position.set(4, 2, 0);
    scene.add(directionalLight);

    // -- Renderer --
    const renderer = new THREE.WebGPURenderer({ antialias: true });
    renderer.setClearColor(0x0a0f1c, 1);
    container.appendChild(renderer.domElement);
    renderer.domElement.style.width = '100%';
    renderer.domElement.style.height = '100%';
    renderer.domElement.style.display = 'block';

    // -- Raycaster for mouse → 3D --
    const raycaster = new THREE.Raycaster();
    const groundPlane = new THREE.Plane(new THREE.Vector3(0, 1, 0), 0);
    const intersectPoint = new THREE.Vector3();

    // -- Particles (40% fewer: ~78K) --
    const count = 2 ** 17; // 131072

    // Physics uniforms
    const attractorMassUniform = uniform(1e7);
    const particleGlobalMassUniform = uniform(1e4);
    const spinningStrengthUniform = uniform(2.75);
    const maxSpeedUniform = uniform(8);
    const velocityDampingUniform = uniform(0.1);
    const scaleUniform = uniform(0.008);
    const boundHalfExtentUniform = uniform(14);
    const colorAUniform = uniform(color('#5900ff'));
    const colorBUniform = uniform(color('#ffa575'));

    // Mouse attractor uniform (position + strength)
    const mousePositionUniform = uniform(new THREE.Vector3(0, 0, 0));
    const mouseStrengthUniform = uniform(0);

    // Static attractors (positions and rotation axes as vec4 arrays)
    const staticPositions = [
      new THREE.Vector4(-3, 0, 0, 0),
      new THREE.Vector4(3, 0, -1.5, 0),
      new THREE.Vector4(0, 1, 3, 0),
    ];
    const staticRotAxes = [
      new THREE.Vector4(0, 1, 0, 0),
      new THREE.Vector4(0, 1, 0, 0),
      new THREE.Vector4(...new THREE.Vector3(1, 0, -0.5).normalize().toArray(), 0),
    ];

    // Storage buffers for positions and velocities
    const positionArray = new THREE.StorageInstancedBufferAttribute(count, 3);
    const velocityArray = new THREE.StorageInstancedBufferAttribute(count, 3);

    const positionStorage = storage(positionArray, 'vec3', count);
    const velocityStorage = storage(velocityArray, 'vec3', count);

    // Attractor data as storage buffers
    const attractorPosData = new Float32Array(3 * 4);
    const attractorRotData = new Float32Array(3 * 4);
    for (let i = 0; i < 3; i++) {
      attractorPosData[i * 4] = staticPositions[i].x;
      attractorPosData[i * 4 + 1] = staticPositions[i].y;
      attractorPosData[i * 4 + 2] = staticPositions[i].z;
      attractorPosData[i * 4 + 3] = 0;
      attractorRotData[i * 4] = staticRotAxes[i].x;
      attractorRotData[i * 4 + 1] = staticRotAxes[i].y;
      attractorRotData[i * 4 + 2] = staticRotAxes[i].z;
      attractorRotData[i * 4 + 3] = 0;
    }
    const attractorPosAttr = new THREE.StorageBufferAttribute(attractorPosData, 4);
    const attractorRotAttr = new THREE.StorageBufferAttribute(attractorRotData, 4);
    const attractorPosStorage = storage(attractorPosAttr, 'vec4', 3).toReadOnly();
    const attractorRotStorage = storage(attractorRotAttr, 'vec4', 3).toReadOnly();

    // -- Init Compute: randomize positions & velocities --
    const initFn = Fn(() => {
      const idx = instanceIndex;
      const seed = float(idx).mul(0.000007629).add(0.5).fract();
      const seed2 = float(idx).mul(0.000013).add(0.3).fract();
      const seed3 = float(idx).mul(0.000019).add(0.7).fract();
      const seed4 = float(idx).mul(0.000029).add(0.1).fract();
      const seed5 = float(idx).mul(0.000037).add(0.9).fract();

      // Spawn particles near the 3 static attractors
      // Each particle picks one attractor based on its index
      const attractorIdx = float(idx).mod(3).floor();
      // Attractor centers: (-3,0,0), (3,0,-1.5), (0,1,3)
      const ax = attractorIdx.equal(0).select(-3, attractorIdx.equal(1).select(3, 0));
      const ay = attractorIdx.equal(0).select(0, attractorIdx.equal(1).select(0, 1));
      const az = attractorIdx.equal(0).select(0, attractorIdx.equal(1).select(-1.5, 3));

      // Spread around attractor with some randomness
      const px = ax.add(sub(seed, 0.5).mul(4));
      const py = ay.add(sub(seed2, 0.5).mul(4));
      const pz = az.add(sub(seed3, 0.5).mul(4));
      positionStorage.element(idx).assign(vec3(px, py, pz));

      const phi = seed4.mul(Math.PI * 2);
      const theta = seed5.mul(2);
      const sinPhi = phi.sin();
      const vx = sinPhi.mul(theta.sin()).mul(0.3);
      const vy = phi.cos().mul(0.3);
      const vz = sinPhi.mul(theta.cos()).mul(0.3);
      velocityStorage.element(idx).assign(vec3(vx, vy, vz));
    });

    const initCompute = initFn().compute(count);

    // -- Update Compute: physics --
    const updateFn = Fn(() => {
      const delta = float(1 / 60);
      const idx = instanceIndex;
      const position = positionStorage.element(idx).toVar();
      const velocity = velocityStorage.element(idx).toVar();

      const massSeed = float(idx).mul(0.000007629).add(0.5).fract();
      const particleMassMultiplier = massSeed.mul(0.75).add(0.25);
      const particleMass = particleMassMultiplier.mul(particleGlobalMassUniform);

      const G = float(6.67e-11);
      const force = vec3(0, 0, 0).toVar();

      // Static attractors
      Loop(3, ({ i }) => {
        const attractorPos = attractorPosStorage.element(i).xyz;
        const attractorRot = attractorRotStorage.element(i).xyz;

        const toAttractor = sub(attractorPos, position);
        const dist = max(length(toAttractor), 0.1);
        const direction = normalize(toAttractor);

        const gravityStrength = div(
          mul(attractorMassUniform, mul(particleMass, G)),
          mul(dist, dist)
        );
        const gravityForce = mul(direction, gravityStrength);
        force.addAssign(gravityForce);

        const spinForce = mul(attractorRot, mul(gravityStrength, spinningStrengthUniform));
        const spinVelocity = cross(spinForce, toAttractor);
        force.addAssign(spinVelocity);
      });

      // Mouse attractor — always compute, scale by strength uniform
      const toMouse = sub(mousePositionUniform, position);
      const mouseDist = max(length(toMouse), 0.1);
      const mouseDir = normalize(toMouse);

      const mouseMass = mul(attractorMassUniform, 1.25); // 5/4 = 1.25x base (was 5x)
      const mouseGravity = div(
        mul(mouseMass, mul(particleMass, G)),
        mul(mouseDist, mouseDist)
      ).mul(mouseStrengthUniform);

      force.addAssign(mul(mouseDir, mouseGravity));

      const mouseRotAxis = vec3(0, 1, 0);
      const mouseSpinForce = mul(mouseRotAxis, mul(mouseGravity, spinningStrengthUniform));
      force.addAssign(cross(mouseSpinForce, toMouse));

      // Velocity update
      velocity.addAssign(mul(force, delta));
      const speed = length(velocity);
      If(speed.greaterThan(maxSpeedUniform), () => {
        velocity.assign(mul(normalize(velocity), maxSpeedUniform));
      });
      velocity.mulAssign(sub(1, velocityDampingUniform));

      // Position update
      position.addAssign(mul(velocity, delta));

      // Toroidal wrapping
      const halfExtent = div(boundHalfExtentUniform, 2);
      position.assign(sub(mod(add(position, halfExtent), boundHalfExtentUniform), halfExtent));

      positionStorage.element(idx).assign(position);
      velocityStorage.element(idx).assign(velocity);
    });

    const updateCompute = updateFn().compute(count);

    // -- Material --
    const material = new THREE.SpriteNodeMaterial({
      blending: THREE.AdditiveBlending,
      depthWrite: false,
    });

    material.positionNode = positionStorage.toAttribute();

    material.colorNode = Fn(() => {
      const vel = velocityStorage.toAttribute();
      const speed = length(vel);
      const colorMix = smoothstep(0, 0.5, div(speed, maxSpeedUniform));
      const finalColor = mix(colorAUniform, colorBUniform, colorMix);
      return vec4(finalColor, 1);
    })();

    material.scaleNode = Fn(() => {
      const massSeed = float(instanceIndex).mul(0.000007629).add(0.5).fract();
      return massSeed.mul(0.75).add(0.25).mul(scaleUniform);
    })();

    // -- Mesh --
    const geometry = new THREE.PlaneGeometry(1, 1);
    const mesh = new THREE.InstancedMesh(geometry, material, count);
    scene.add(mesh);

    // -- Init --
    async function init() {
      await renderer.init();
      if (disposed) return;
      initialized = true;
      renderer.compute(initCompute);
      renderer.setAnimationLoop(animate);
    }

    // -- Animate --
    function animate() {
      if (!isVisible) return;

      // Lerp mouse attractor strength
      const targetStrength = mouseOnCanvas ? 1 : 0;
      mouseAttractorStrength += (targetStrength - mouseAttractorStrength) * 0.05;
      mouseStrengthUniform.value = mouseAttractorStrength;

      // Update mouse world position via raycasting
      if (mouseOnCanvas) {
        raycaster.setFromCamera(mouse, camera);
        const hit = raycaster.ray.intersectPlane(groundPlane, intersectPoint);
        if (hit) {
          mouseWorld.lerp(intersectPoint, 0.15);
          (mousePositionUniform.value as THREE.Vector3).copy(mouseWorld);
        }
      }

      // Scroll → zoom in (camera moves closer)
      const scrollRatio = Math.min(currentScrollY / 800, 1); // 0 to 1 over 800px scroll
      const targetZoom = scrollRatio;
      currentCameraAngle += (targetZoom - currentCameraAngle) * 0.15;
      const zoomFactor = 1 - currentCameraAngle * 0.7; // zooms in up to 70%
      camera.position.set(0, 3 * zoomFactor, 5 * zoomFactor);
      camera.lookAt(0, 0, 0);

      renderer.compute(updateCompute);
      renderer.render(scene, camera);
    }

    // -- Events: listen on window so hero content overlay doesn't block --
    function onMouseMove(e: MouseEvent) {
      if (!container) return;
      const rect = container.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      // Check if mouse is within the container bounds
      if (x >= 0 && x <= rect.width && y >= 0 && y <= rect.height) {
        mouse.x = (x / rect.width) * 2 - 1;
        mouse.y = -(y / rect.height) * 2 + 1;
        mouseOnCanvas = true;
      } else {
        mouseOnCanvas = false;
      }
    }

    function onMouseLeave() {
      mouseOnCanvas = false;
    }

    function onScroll() {
      currentScrollY = window.scrollY;
    }

    window.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseleave', onMouseLeave);
    window.addEventListener('scroll', onScroll, { passive: true });

    // -- Resize --
    function handleResize(entries: ResizeObserverEntry[]) {
      const entry = entries[0];
      if (!entry) return;
      const { width, height } = entry.contentRect;
      if (width === 0 || height === 0) return;
      renderer.setSize(width, height, false);
      renderer.setPixelRatio(window.devicePixelRatio);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
    }

    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(container);

    // -- Visibility: pause when hero scrolls out of view --
    const visibilityObserver = new IntersectionObserver(
      ([entry]) => { isVisible = entry.isIntersecting; },
      { threshold: 0 }
    );
    visibilityObserver.observe(container);

    const { clientWidth, clientHeight } = container;
    if (clientWidth > 0 && clientHeight > 0) {
      renderer.setSize(clientWidth, clientHeight, false);
      renderer.setPixelRatio(window.devicePixelRatio);
      camera.aspect = clientWidth / clientHeight;
      camera.updateProjectionMatrix();
    }

    init();

    return () => {
      disposed = true;
      renderer.setAnimationLoop(null);
      resizeObserver.disconnect();
      visibilityObserver.disconnect();
      window.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseleave', onMouseLeave);
      window.removeEventListener('scroll', onScroll);
      if (initialized) {
        renderer.dispose();
        geometry.dispose();
        material.dispose();
      }
      if (renderer.domElement.parentNode) {
        renderer.domElement.parentNode.removeChild(renderer.domElement);
      }
    };
  }, []);

  return (
    <div
      ref={containerRef}
      className="absolute inset-0 w-full h-full"
      style={{ opacity: 0.85 }}
    />
  );
}
