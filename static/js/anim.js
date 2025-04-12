/******************************************************
 * CONFIG
 ******************************************************/
// We'll let fraction go from 0..1 as user scrolls 0..2000
let maxScroll;
// Both letters eventually share orbit radius=10 at fraction=1
const finalOrbitRad = 10;
// "v" rotation 0..-90
const maxVRotation = -90;
// tilt the container 0..-10 deg
const maxTilt = -15;
// baseSpin for continuous rotation (now a multiplier)
const onScrollSpinMultiplier = 0.1; // Controls spin speed during scroll
const restingSpinMultiplier = 2; // Controls spin speed when not scrolling

// Bits drift left 50px & scale 1->0.5 in 0.5s
const BIT_LIFETIME = 500;
const BIT_DISTANCE = 500;
const startScale = 1.0;
const endScale = 0.1;

// New config parameter for bit drift speed (multiplier)
const bitTrailDriftSpeed = 0.05; // Default speed (1x). Higher values are faster.

// New config parameter for bits per propeller per second
const bitsPerPropellerPerSecond = 2; // Adjust this value

// wind sway
const windFrequency = 0.05;
const windXAmplitude = 6;
const windYAmplitude = 3;

let planeContainer, lettersRow;
let pSpan, gSpan, vSpan;
let orbitP, orbitG;

// We'll define a single orbit center = midpoint of p & g
let centerX = 0, centerY = 0;

// We'll store p/g's initial distance & angle from that center
// so at fraction=0, they remain where they are (no snap).
let pInitDist = 0, gInitDist = 0;
let pInitAngle = 0, gInitAngle = 0;

// We'll keep a separate orbitAngle that increments each frame if fraction>0
// so p & g revolve around the center.
let orbitAngle = 0;
let orbitCreated = false;
let frameCount = 0;

// bits array (will only store data, not DOM elements directly)
const bits = [];

// Variables to track scroll speed
let previousScrollY = 0;
let previousTime = 0;
let animTimeline;

window.addEventListener('DOMContentLoaded', () => {
  // Calculate maxScroll based on page height
  maxScroll = Math.max(
    document.body.scrollHeight,
    document.documentElement.scrollHeight,
    document.body.offsetHeight,
    document.documentElement.offsetHeight,
    document.body.clientHeight,
    document.documentElement.clientHeight
  ) - window.innerHeight;

  planeContainer = document.getElementById('planeContainer');
  lettersRow = document.getElementById('lettersRow');
  pSpan = document.getElementById('pSpan');
  gSpan = document.getElementById('gSpan');
  vSpan = document.getElementById('vSpan');

  // Create absolutely positioned orbit p & g
  orbitP = document.createElement('div');
  orbitP.className = 'orbit-p';
  orbitP.textContent = 'p';
  lettersRow.appendChild(orbitP);

  orbitG = document.createElement('div');
  orbitG.className = 'orbit-g';
  orbitG.textContent = 'g';
  lettersRow.appendChild(orbitG);

  // measure pSpan position
  const pRect = pSpan.getBoundingClientRect();
  // measure gSpan position
  const gRect = gSpan.getBoundingClientRect();
  // measure row's top-left
  const rowRect = lettersRow.getBoundingClientRect();

  // convert absolute => local coords
  const pX = pRect.left - rowRect.left;
  const pY = pRect.top - rowRect.top;
  const gX = gRect.left - rowRect.left;
  const gY = gRect.top - rowRect.top;

  // Orbit center = midpoint of p & g
  centerX = (pX + gX) / 2;
  centerY = (pY + gY) / 2;

  // p's initial offset from center
  let dxP = pX - centerX;
  let dyP = pY - centerY;
  pInitDist = Math.sqrt(dxP * dxP + dyP * dyP);
  pInitAngle = Math.atan2(dyP, dxP) * (180 / Math.PI);

  // g's initial offset from center
  let dxG = gX - centerX;
  let dyG = gY - centerY;
  gInitDist = Math.sqrt(dxG * dxG + dyG * dyG);
  gInitAngle = Math.atan2(dyG, dxG) * (180 / Math.PI);

  // Initialize previous time
  previousTime = performance.now();

  // Force document height to allow for full scroll range
  // document.body.style.minHeight = '3000px';
});

window.addEventListener('resize', () => {
  maxScroll = Math.max( 
    document.body.scrollHeight,
    document.documentElement.scrollHeight,
    document.body.offsetHeight,
    document.documentElement.offsetHeight,
    document.body.clientHeight,
    document.documentElement.clientHeight
  ) - window.innerHeight;
});

window.onload = async () => {
  initializeTimeline();
};

function initializeTimeline() {
  animTimeline = anime.timeline({
    autoplay: false,
  })
    .add({ // Phase 1: 0% to 30% scroll (Tilt)
      targets: planeContainer,
      rotate: maxTilt,
      easing: 'easeInOutQuad',
      duration: maxScroll * 0.3,
    }, 0)
    .add({ // Phase 1: 0% to 30% scroll (vSpan Rotation)
      targets: vSpan,
      rotate: maxVRotation,
      easing: 'easeInOutQuad',
      duration: maxScroll * 0.3,// Keep the final rotation value
    }, 0);

  // Start the animation loop
  update();
}

/******************************************************
 * ANIMATION LOOP
 ******************************************************/
function update() {
  let scrollY = window.scrollY;
  if (scrollY < 0) scrollY = 0;
  if (scrollY > maxScroll) scrollY = maxScroll;
  let fraction = scrollY / maxScroll;

  // Calculate scroll speed (for propeller spin)
  const currentTime = performance.now();
  const deltaTime = currentTime - previousTime;
  const deltaScroll = Math.abs(scrollY - previousScrollY);
  const scrollSpeed = deltaTime > 0 ? deltaScroll / deltaTime : 0; // pixels per millisecond
  previousScrollY = scrollY;
  previousTime = currentTime;

  // Adjust spin based on scroll speed
  let dynamicBaseSpin = onScrollSpinMultiplier * scrollSpeed * 60;
  // Ensure dynamicBaseSpin is always at least restingSpinMultiplier
  if (dynamicBaseSpin < restingSpinMultiplier) {
    dynamicBaseSpin = restingSpinMultiplier;
  }

  // Propeller Rotation (always on when fraction > 0)
  if (fraction > 0) {
    orbitAngle += dynamicBaseSpin;
    frameCount++;

    // Keep orbit radius growing until fraction is 1
    let pRadius = pInitDist + (finalOrbitRad - pInitDist) * Math.min(1, fraction);
    let gRadius = gInitDist + (finalOrbitRad - gInitDist) * Math.min(1, fraction);

    let pA = (pInitAngle + orbitAngle) * Math.PI / 180;
    let xP = centerX + Math.cos(pA) * pRadius;
    let yP = centerY + Math.sin(pA) * pRadius;

    anime({
      targets: orbitP,
      left: xP,
      top: yP,
      duration: 0,
      easing: 'linear'
    });

    let gA = (gInitAngle + orbitAngle) * Math.PI / 180;
    let xG = centerX + Math.cos(gA) * gRadius;
    let yG = centerY + Math.sin(gA) * gRadius;

    anime({
      targets: orbitG,
      left: xG,
      top: yG,
      duration: 0,
      easing: 'linear'
    });
  } else {
    orbitAngle = 0;
    frameCount = 0;
  }

  // Show/hide orbit p/g
  if (fraction > 0) {
    if (!orbitCreated) {
      pSpan.style.opacity = 0;
      gSpan.style.opacity = 0;
      orbitP.style.opacity = 1;
      orbitG.style.opacity = 1;
      orbitCreated = true;
    }
  } else {
    pSpan.style.opacity = 1;
    gSpan.style.opacity = 1;
    orbitP.style.opacity = 0;
    orbitG.style.opacity = 0;
    orbitCreated = false;
  }

  // Wind Sway (always on when fraction > 0)
  let offsetX = 0, offsetY = 0;
  if (fraction > 0) {
    offsetX = Math.sin(frameCount * windFrequency) * windXAmplitude;
    offsetY = Math.cos(frameCount * windFrequency) * windYAmplitude;
    anime({
      targets: planeContainer,
      rotate: maxTilt,
      translateX: offsetX,
      translateY: offsetY,
      duration: 0,
      easing: 'linear'
    });
  }

  // Bit Trails (always on when fraction > 0)
  const bitsPerFrame = bitsPerPropellerPerSecond / 60; // Approximate bits per frame
  if (fraction > 0 && Math.random() < bitsPerFrame * 2) { // x2 for both propellers
    const pRectOrbit = orbitP.getBoundingClientRect();
    const gRectOrbit = orbitG.getBoundingClientRect();
    const rowRectLocal = lettersRow.getBoundingClientRect();
    spawnBit((pRectOrbit.left - rowRectLocal.left + pRectOrbit.width) / 2, (pRectOrbit.top - rowRectLocal.top + pRectOrbit.height) / 2);
    spawnBit((gRectOrbit.left - rowRectLocal.left + gRectOrbit.width) / 2, (gRectOrbit.top - rowRectLocal.top + gRectOrbit.height) / 2);
  }
  updateBits(fraction);

  // Control the Anime.js timeline progress
  if (animTimeline) {
    animTimeline.seek(scrollY);
  }

  requestAnimationFrame(update);
}

/******************************************************
 * EASING: easeOutQuad(t) = 1 - (1-t)^2
 ******************************************************/
function easeOutQuad(t) {
  return 1 - (1 - t) * (1 - t);
}

/******************************************************
 * updateBits => drift left 50px & scale 1->0.5 in 0.5s
 ******************************************************/
function updateBits(fraction) {
  const now = performance.now();
  for (let i = bits.length - 1; i >= 0; i--) {
    const bitData = bits[i];
    const bitEl = document.getElementById(bitData.id);
    if (!bitEl) {
      bits.splice(i, 1);
      continue;
    }

    const lifetime = now - bitData.spawnTime;
    const progress = Math.min(1, (lifetime / BIT_LIFETIME) * bitTrailDriftSpeed);

    if (progress > 1) {
      bits.splice(i, 1);
      bitEl.remove();
      continue;
    }
    const eased = easeOutQuad(progress);

    const targetX = bitData.xSpawn - BIT_DISTANCE * eased;
    const targetScale = startScale + (endScale - startScale) * eased;
    const targetOpacity = 1 - progress;

    anime({
      targets: bitEl,
      translateX: targetX - bitData.xSpawn,
      scale: targetScale,
      opacity: targetOpacity,
      duration: 0,
      easing: 'linear'
    });
  }
}

/******************************************************
 * spawnBit => put a grey bit under p/g,
 * then let it move left 50px & scale 1->0.5 in 0.5s.
 ******************************************************/
function spawnBit(x, y) {
  const bitVal = Math.random() < 0.5 ? '0' : '1';
  const bitEl = document.createElement('div');
  bitEl.className = 'bit-trail';
  bitEl.textContent = bitVal;
  const bitId = `bit-${performance.now()}-${Math.random()}`;
  bitEl.id = bitId;

  // Place at (x, y)
  bitEl.style.left = x + 'px';
  bitEl.style.top = y + 'px';

  lettersRow.appendChild(bitEl);

  const bitData = {
    id: bitId,
    xSpawn: x,
    ySpawn: y,
    spawnTime: performance.now()
  };
  bits.push(bitData);

  anime({
    targets: bitEl,
    translateX: -BIT_DISTANCE,
    scale: endScale,
    opacity: 0,
    duration: BIT_LIFETIME * (1 / bitTrailDriftSpeed),
    easing: 'easeOutQuad',
    complete: function (anim) {
      bitEl.remove();
      const index = bits.findIndex(bit => bit.id === bitId);
      if (index > -1) {
        bits.splice(index, 1);
      }
    }
  });
}

function loadFullViewportSvg(svgUrl, callback) {
  fetch(svgUrl)
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return response.text();
    })
    .then(svgText => {
      // Create a temporary div to hold the SVG content
      const tempDiv = document.createElement('div');
      tempDiv.innerHTML = svgText;

      // Get the actual SVG element from the parsed content
      const svgElement = tempDiv.querySelector('svg');

      if (!svgElement) {
        throw new Error('Could not find SVG element in the loaded content.');
      }

      // Position SVG at 30% from top
      svgElement.style.position = 'fixed';
      svgElement.style.top = `${maxScroll * 0.3}px`;

      // svgElement.setAttribute('width', `${maxScroll}px`);
      // svgElement.setAttribute('preserveAspectRatio', 'none');

      // Append the SVG to the body
      document.body.appendChild(svgElement);

      // Ensure the body and HTML take up the full viewport to avoid scrollbars
      document.body.style.margin = '0';
      // document.body.style.overflow = 'hidden';
      // document.documentElement.style.height = `${maxScroll}px`;
      // document.body.style.height = '100%';

      // Execute the callback
      if (callback && typeof callback === 'function') {
        callback(svgElement, null);
      }
    })
    .catch(error => {
      console.error('Error loading SVG:', error);
      // You might want to handle the error in the callback as well
      if (callback && typeof callback === 'function') {
        callback(null, error); // Pass the error to the callback
      }
    });
}
