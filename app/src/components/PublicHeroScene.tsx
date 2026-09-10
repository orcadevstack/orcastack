import { useEffect, useRef } from 'react';
import * as THREE from 'three';

export function PublicHeroScene() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = containerRef.current;
    if (!host) return;

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(42, 1, 0.1, 100);
    camera.position.set(0, 0.2, 8.5);

    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true, powerPreference: 'high-performance', preserveDrawingBuffer: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
    renderer.outputColorSpace = THREE.SRGBColorSpace;
    host.appendChild(renderer.domElement);

    const group = new THREE.Group();
    scene.add(group);

    const core = new THREE.Mesh(
      new THREE.IcosahedronGeometry(1.55, 2),
      new THREE.MeshStandardMaterial({ color: 0xfc6d26, metalness: 0.45, roughness: 0.28, emissive: 0x351006, emissiveIntensity: 0.7 }),
    );
    group.add(core);

    const wire = new THREE.Mesh(
      new THREE.IcosahedronGeometry(2.15, 1),
      new THREE.MeshBasicMaterial({ color: 0xffb48a, wireframe: true, transparent: true, opacity: 0.22 }),
    );
    group.add(wire);

    const ringMaterial = new THREE.MeshBasicMaterial({ color: 0xe8edf5, transparent: true, opacity: 0.2, side: THREE.DoubleSide });
    for (const [index, radius] of [2.75, 3.3, 3.85].entries()) {
      const ring = new THREE.Mesh(new THREE.RingGeometry(radius, radius + 0.018, 128), ringMaterial.clone());
      ring.rotation.x = Math.PI / (2.5 + index * 0.4);
      ring.rotation.y = index * 0.58;
      group.add(ring);
    }

    const points = new Float32Array(360 * 3);
    for (let index = 0; index < 360; index += 1) {
      const radius = 3.2 + Math.random() * 3.8;
      const angle = Math.random() * Math.PI * 2;
      const height = (Math.random() - 0.5) * 5.5;
      points[index * 3] = Math.cos(angle) * radius;
      points[index * 3 + 1] = height;
      points[index * 3 + 2] = Math.sin(angle) * radius;
    }
    const pointGeometry = new THREE.BufferGeometry();
    pointGeometry.setAttribute('position', new THREE.BufferAttribute(points, 3));
    const particleField = new THREE.Points(pointGeometry, new THREE.PointsMaterial({ color: 0xf4f7fb, size: 0.025, transparent: true, opacity: 0.42 }));
    scene.add(particleField);

    scene.add(new THREE.HemisphereLight(0xffd8c2, 0x101b2c, 1.5));
    const keyLight = new THREE.PointLight(0xfc6d26, 28, 18);
    keyLight.position.set(3, 3, 4);
    scene.add(keyLight);
    const rimLight = new THREE.PointLight(0x7dd3fc, 18, 18);
    rimLight.position.set(-4, -2, 3);
    scene.add(rimLight);

    const pointer = new THREE.Vector2();
    const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    function onPointerMove(event: PointerEvent) {
      pointer.x = (event.clientX / window.innerWidth - 0.5) * 0.5;
      pointer.y = (event.clientY / window.innerHeight - 0.5) * 0.35;
    }
    window.addEventListener('pointermove', onPointerMove, { passive: true });

    function resize() {
      const width = host!.clientWidth;
      const height = host!.clientHeight;
      camera.aspect = width / Math.max(height, 1);
      camera.updateProjectionMatrix();
      renderer.setSize(width, height, false);
    }
    const resizeObserver = new ResizeObserver(resize);
    resizeObserver.observe(host);
    resize();

    let frame = 0;
    function render(timestamp = 0) {
      const elapsed = timestamp / 1000;
      if (!reduceMotion) {
        group.rotation.y = elapsed * 0.12 + pointer.x;
        group.rotation.x += (pointer.y - group.rotation.x) * 0.025;
        wire.rotation.z = elapsed * -0.08;
        particleField.rotation.y = elapsed * 0.018;
        core.position.y = Math.sin(elapsed * 0.8) * 0.12;
      }
      renderer.render(scene, camera);
      frame = window.requestAnimationFrame(render);
    }
    render();

    return () => {
      window.cancelAnimationFrame(frame);
      window.removeEventListener('pointermove', onPointerMove);
      resizeObserver.disconnect();
      pointGeometry.dispose();
      core.geometry.dispose();
      wire.geometry.dispose();
      core.material.dispose();
      wire.material.dispose();
      group.children.forEach((child) => {
        if (child instanceof THREE.Mesh) {
          child.geometry.dispose();
          if (child.material instanceof THREE.Material) child.material.dispose();
        }
      });
      renderer.dispose();
      renderer.domElement.remove();
    };
  }, []);

  return <div aria-hidden="true" className="public-hero-scene" ref={containerRef} />;
}
