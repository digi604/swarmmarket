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

// Attractor configs shared by GPU and CPU paths
const ATTRACTORS = [
  { pos: [-3, 0, 0], rot: [0, 1, 0] },
  { pos: [3, 0, -1.5], rot: [0, 1, 0] },
  { pos: [0, 1, 3], rot: [...new THREE.Vector3(1, 0, -0.5).normalize().toArray()] },
] as const;

const P = { mass: 1e7, pMass: 1e4, spin: 2.75, maxSpd: 8, damp: 0.1, bounds: 14, G: 6.67e-11 };

export function Particles() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let disposed = false;
    let initialized = false;
    let isVisible = true;

    const mouse = new THREE.Vector2(0, 0);
    let mouseOnCanvas = false;
    let mouseAttractorStrength = 0;
    const mouseWorld = new THREE.Vector3(0, 0, 0);
    let currentScrollY = window.scrollY;
    let currentCameraAngle = 0;

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(60, 1, 0.1, 100);
    camera.position.set(0, 3, 5);
    camera.lookAt(0, 0, 0);

    scene.add(new THREE.AmbientLight('#ffffff', 0.5));
    const dirLight = new THREE.DirectionalLight('#ffffff', 1.5);
    dirLight.position.set(4, 2, 0);
    scene.add(dirLight);

    const renderer = new THREE.WebGPURenderer({ antialias: true });
    renderer.setClearColor(0x0a0f1c, 1);
    container.appendChild(renderer.domElement);
    renderer.domElement.style.width = '100%';
    renderer.domElement.style.height = '100%';
    renderer.domElement.style.display = 'block';

    const raycaster = new THREE.Raycaster();
    const groundPlane = new THREE.Plane(new THREE.Vector3(0, 1, 0), 0);
    const intersectPoint = new THREE.Vector3();

    let geo: THREE.PlaneGeometry | null = null;
    let mat: THREE.SpriteNodeMaterial | null = null;

    function updateSharedState() {
      const target = mouseOnCanvas ? 1 : 0;
      mouseAttractorStrength += (target - mouseAttractorStrength) * 0.05;
      if (mouseOnCanvas) {
        raycaster.setFromCamera(mouse, camera);
        const hit = raycaster.ray.intersectPlane(groundPlane, intersectPoint);
        if (hit) mouseWorld.lerp(intersectPoint, 0.15);
      }
      const scrollRatio = Math.min(currentScrollY / 800, 1);
      currentCameraAngle += (scrollRatio - currentCameraAngle) * 0.15;
      const z = 1 - currentCameraAngle * 0.7;
      camera.position.set(0, 3 * z, 5 * z);
      camera.lookAt(0, 0, 0);
    }

    // ── WebGPU path: compute shaders, 131K particles ──

    function setupGPU() {
      const count = 2 ** 17; // 131072

      const attractorMassUniform = uniform(P.mass);
      const particleGlobalMassUniform = uniform(P.pMass);
      const spinningStrengthUniform = uniform(P.spin);
      const maxSpeedUniform = uniform(P.maxSpd);
      const velocityDampingUniform = uniform(P.damp);
      const scaleUniform = uniform(0.008);
      const boundHalfExtentUniform = uniform(P.bounds);
      const colorAUniform = uniform(color('#5900ff'));
      const colorBUniform = uniform(color('#ffa575'));
      const mousePositionUniform = uniform(new THREE.Vector3(0, 0, 0));
      const mouseStrengthUniform = uniform(0);

      const attractorPosData = new Float32Array(12);
      const attractorRotData = new Float32Array(12);
      for (let i = 0; i < 3; i++) {
        attractorPosData.set(ATTRACTORS[i].pos, i * 4);
        attractorRotData.set(ATTRACTORS[i].rot, i * 4);
      }
      const attractorPosStorage = storage(new THREE.StorageBufferAttribute(attractorPosData, 4), 'vec4', 3).toReadOnly();
      const attractorRotStorage = storage(new THREE.StorageBufferAttribute(attractorRotData, 4), 'vec4', 3).toReadOnly();

      const positionArray = new THREE.StorageInstancedBufferAttribute(count, 3);
      const velocityArray = new THREE.StorageInstancedBufferAttribute(count, 3);
      const positionStorage = storage(positionArray, 'vec3', count);
      const velocityStorage = storage(velocityArray, 'vec3', count);

      const initCompute = Fn(() => {
        const idx = instanceIndex;
        const s1 = float(idx).mul(0.000007629).add(0.5).fract();
        const s2 = float(idx).mul(0.000013).add(0.3).fract();
        const s3 = float(idx).mul(0.000019).add(0.7).fract();
        const s4 = float(idx).mul(0.000029).add(0.1).fract();
        const s5 = float(idx).mul(0.000037).add(0.9).fract();
        const ai = float(idx).mod(3).floor();
        const ax = ai.equal(0).select(-3, ai.equal(1).select(3, 0));
        const ay = ai.equal(0).select(0, ai.equal(1).select(0, 1));
        const az = ai.equal(0).select(0, ai.equal(1).select(-1.5, 3));
        positionStorage.element(idx).assign(vec3(ax.add(sub(s1, 0.5).mul(4)), ay.add(sub(s2, 0.5).mul(4)), az.add(sub(s3, 0.5).mul(4))));
        const phi = s4.mul(Math.PI * 2);
        const sinPhi = phi.sin();
        velocityStorage.element(idx).assign(vec3(sinPhi.mul(s5.mul(2).sin()).mul(0.3), phi.cos().mul(0.3), sinPhi.mul(s5.mul(2).cos()).mul(0.3)));
      })().compute(count);

      const updateCompute = Fn(() => {
        const delta = float(1 / 60);
        const idx = instanceIndex;
        const position = positionStorage.element(idx).toVar();
        const velocity = velocityStorage.element(idx).toVar();
        const pMass = float(idx).mul(0.000007629).add(0.5).fract().mul(0.75).add(0.25).mul(particleGlobalMassUniform);
        const Gc = float(P.G);
        const force = vec3(0, 0, 0).toVar();
        Loop(3, ({ i }) => {
          const aPos = attractorPosStorage.element(i).xyz;
          const aRot = attractorRotStorage.element(i).xyz;
          const toA = sub(aPos, position);
          const dist = max(length(toA), 0.1);
          const gStr = div(mul(attractorMassUniform, mul(pMass, Gc)), mul(dist, dist));
          force.addAssign(mul(normalize(toA), gStr));
          force.addAssign(cross(mul(aRot, mul(gStr, spinningStrengthUniform)), toA));
        });
        const toM = sub(mousePositionUniform, position);
        const mDist = max(length(toM), 0.1);
        const mGrav = div(mul(mul(attractorMassUniform, 1.25), mul(pMass, Gc)), mul(mDist, mDist)).mul(mouseStrengthUniform);
        force.addAssign(mul(normalize(toM), mGrav));
        force.addAssign(cross(mul(vec3(0, 1, 0), mul(mGrav, spinningStrengthUniform)), toM));
        velocity.addAssign(mul(force, delta));
        If(length(velocity).greaterThan(maxSpeedUniform), () => { velocity.assign(mul(normalize(velocity), maxSpeedUniform)); });
        velocity.mulAssign(sub(1, velocityDampingUniform));
        position.addAssign(mul(velocity, delta));
        const half = div(boundHalfExtentUniform, 2);
        position.assign(sub(mod(add(position, half), boundHalfExtentUniform), half));
        positionStorage.element(idx).assign(position);
        velocityStorage.element(idx).assign(velocity);
      })().compute(count);

      mat = new THREE.SpriteNodeMaterial({ blending: THREE.AdditiveBlending, depthWrite: false });
      mat.positionNode = positionStorage.toAttribute();
      mat.colorNode = Fn(() => {
        const spd = length(velocityStorage.toAttribute());
        return vec4(mix(colorAUniform, colorBUniform, smoothstep(0, 0.5, div(spd, maxSpeedUniform))), 1);
      })();
      mat.scaleNode = Fn(() => float(instanceIndex).mul(0.000007629).add(0.5).fract().mul(0.75).add(0.25).mul(scaleUniform))();

      geo = new THREE.PlaneGeometry(1, 1);
      scene.add(new THREE.InstancedMesh(geo, mat, count));

      renderer.compute(initCompute);
      renderer.setAnimationLoop(() => {
        if (!isVisible) return;
        updateSharedState();
        mouseStrengthUniform.value = mouseAttractorStrength;
        (mousePositionUniform.value as THREE.Vector3).copy(mouseWorld);
        renderer.compute(updateCompute);
        renderer.render(scene, camera);
      });
    }

    // ── CPU fallback for WebGL2 backend (avoids broken transform feedback on NVIDIA Linux) ──

    function setupCPU() {
      const count = 2 ** 14; // 16384
      const posAttr = new THREE.StorageInstancedBufferAttribute(count, 3);
      const velAttr = new THREE.StorageInstancedBufferAttribute(count, 3);
      const pos = posAttr.array as Float32Array;
      const vel = velAttr.array as Float32Array;

      for (let idx = 0; idx < count; idx++) {
        const s1 = (idx * 0.000007629 + 0.5) % 1;
        const s2 = (idx * 0.000013 + 0.3) % 1;
        const s3 = (idx * 0.000019 + 0.7) % 1;
        const s4 = (idx * 0.000029 + 0.1) % 1;
        const s5 = (idx * 0.000037 + 0.9) % 1;
        const a = ATTRACTORS[idx % 3];
        const i3 = idx * 3;
        pos[i3] = a.pos[0] + (s1 - 0.5) * 4;
        pos[i3 + 1] = a.pos[1] + (s2 - 0.5) * 4;
        pos[i3 + 2] = a.pos[2] + (s3 - 0.5) * 4;
        const phi = s4 * Math.PI * 2;
        const sinPhi = Math.sin(phi);
        vel[i3] = sinPhi * Math.sin(s5 * 2) * 0.3;
        vel[i3 + 1] = Math.cos(phi) * 0.3;
        vel[i3 + 2] = sinPhi * Math.cos(s5 * 2) * 0.3;
      }

      const posS = storage(posAttr, 'vec3', count);
      const velS = storage(velAttr, 'vec3', count);
      const maxSpeedU = uniform(P.maxSpd);
      const colorAU = uniform(color('#5900ff'));
      const colorBU = uniform(color('#ffa575'));
      const scaleU = uniform(0.008);

      mat = new THREE.SpriteNodeMaterial({ blending: THREE.AdditiveBlending, depthWrite: false });
      mat.positionNode = posS.toAttribute();
      mat.colorNode = Fn(() => {
        const spd = length(velS.toAttribute());
        return vec4(mix(colorAU, colorBU, smoothstep(0, 0.5, div(spd, maxSpeedU))), 1);
      })();
      mat.scaleNode = Fn(() => float(instanceIndex).mul(0.000007629).add(0.5).fract().mul(0.75).add(0.25).mul(scaleU))();

      geo = new THREE.PlaneGeometry(1, 1);
      scene.add(new THREE.InstancedMesh(geo, mat, count));

      renderer.setAnimationLoop(() => {
        if (!isVisible) return;
        updateSharedState();
        cpuPhysics(pos, vel, count);
        posAttr.needsUpdate = true;
        velAttr.needsUpdate = true;
        renderer.render(scene, camera);
      });
    }

    function cpuPhysics(pos: Float32Array, vel: Float32Array, count: number) {
      const dt = 1 / 60;
      const mouseStr = mouseAttractorStrength;
      const half = P.bounds / 2;
      const dampF = 1 - P.damp;

      for (let idx = 0; idx < count; idx++) {
        const i3 = idx * 3;
        let px = pos[i3], py = pos[i3 + 1], pz = pos[i3 + 2];
        let vx = vel[i3], vy = vel[i3 + 1], vz = vel[i3 + 2];
        const pMass = ((idx * 0.000007629 + 0.5) % 1 * 0.75 + 0.25) * P.pMass;
        let fx = 0, fy = 0, fz = 0;

        for (let a = 0; a < 3; a++) {
          const ap = ATTRACTORS[a].pos, ar = ATTRACTORS[a].rot;
          const dx = ap[0] - px, dy = ap[1] - py, dz = ap[2] - pz;
          const dist = Math.max(Math.sqrt(dx * dx + dy * dy + dz * dz), 0.1);
          const inv = 1 / dist;
          const gStr = (P.mass * pMass * P.G) / (dist * dist);
          fx += dx * inv * gStr; fy += dy * inv * gStr; fz += dz * inv * gStr;
          const sf = gStr * P.spin;
          // cross(rot * sf, toAttractor)
          fx += ar[1] * sf * dz - ar[2] * sf * dy;
          fy += ar[2] * sf * dx - ar[0] * sf * dz;
          fz += ar[0] * sf * dy - ar[1] * sf * dx;
        }

        if (mouseStr > 0.001) {
          const mx = mouseWorld.x - px, my = mouseWorld.y - py, mz = mouseWorld.z - pz;
          const md = Math.max(Math.sqrt(mx * mx + my * my + mz * mz), 0.1);
          const mg = (P.mass * 1.25 * pMass * P.G) / (md * md) * mouseStr;
          const mi = 1 / md;
          fx += mx * mi * mg; fy += my * mi * mg; fz += mz * mi * mg;
          const ms = mg * P.spin;
          // cross((0, ms, 0), (mx, my, mz))
          fx += ms * mz; fz -= ms * mx;
        }

        vx += fx * dt; vy += fy * dt; vz += fz * dt;
        const spd = Math.sqrt(vx * vx + vy * vy + vz * vz);
        if (spd > P.maxSpd) { const s = P.maxSpd / spd; vx *= s; vy *= s; vz *= s; }
        vx *= dampF; vy *= dampF; vz *= dampF;

        px += vx * dt; py += vy * dt; pz += vz * dt;
        // Toroidal wrapping: (pos + half) mod bounds - half
        px = (((px + half) % P.bounds) + P.bounds) % P.bounds - half;
        py = (((py + half) % P.bounds) + P.bounds) % P.bounds - half;
        pz = (((pz + half) % P.bounds) + P.bounds) % P.bounds - half;

        pos[i3] = px; pos[i3 + 1] = py; pos[i3 + 2] = pz;
        vel[i3] = vx; vel[i3 + 1] = vy; vel[i3 + 2] = vz;
      }
    }

    async function init() {
      await renderer.init();
      if (disposed) return;
      initialized = true;

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const useGPU = !!(renderer as any).backend?.isWebGPUBackend;
      if (useGPU) {
        setupGPU();
      } else {
        setupCPU();
      }
    }

    // ── Events ──

    function onMouseMove(e: MouseEvent) {
      if (!container) return;
      const rect = container.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      if (x >= 0 && x <= rect.width && y >= 0 && y <= rect.height) {
        mouse.x = (x / rect.width) * 2 - 1;
        mouse.y = -(y / rect.height) * 2 + 1;
        mouseOnCanvas = true;
      } else {
        mouseOnCanvas = false;
      }
    }

    function onMouseLeave() { mouseOnCanvas = false; }
    function onScroll() { currentScrollY = window.scrollY; }

    window.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseleave', onMouseLeave);
    window.addEventListener('scroll', onScroll, { passive: true });

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
        scene.remove(mesh);
        geometry.dispose();
        material.dispose();
        renderer.dispose();
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
