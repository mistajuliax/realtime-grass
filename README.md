# realtime-grass

![cover](/assets/images/github/cover.png)

This project is my attempt at Eddie Lee's [real-time grass](http://www.eddietree.com/#/grass/) demo, which was part of his master thesis. The demo consists of an infinite terrain covered in grass that waves in the wind that is exerted by the movement of the camera. Each grass blade is actual geometry as opposed to other approaches that use billboards to show grass patches. The goal is to have a decently fast simulation of natural looking grass.

To achieve this performance the load that is passed to the GPU on each frame has to be minimal. In addition level-of-detail (LOD) techniques are employed to adapt the detail of the grass blades depending on the distance to the camera.

## Skybox

Skybox **TropicalSunnyDay[Back|Down|Front|Left|Right|Up]** is taken from the project [skybox - a player skybox mod](http://minetest.daconcepts.com/my-main-mod-archive/sofars_mods/skybox/textures/) and renamed into **day/[back|bottom|front|left|right|top]**.

SkyboxSet by Heiko Irrgang ( http://gamvas.com ) is licensed under
the Creative Commons Attribution-ShareAlike 3.0 Unported License.
Based on a work at http://93i.de.


## Requirements
This project requires a GPU with OpenGL 4.3+ support.

The following dependencies depend on cgo. To make them work under Windows a compatible version of **mingw** is necessary. Information can be found [here](https://github.com/go-gl/glfw/issues/91). In my case I used *x86_64-7.2.0-posix-seh-rt_v5-rev1*. After installing the right version of **mingw** you can continue by installing the dependencies that follow next.

This project depends on **glfw** for creating a window and providing a rendering context, **go-gl/gl** for providing bindings to OpenGL and **go-gl/mathgl** provides vector and matrix math for OpenGL.
```
go get -u github.com/go-gl/glfw/v3.2/glfw
go get -u github.com/go-gl/gl/v4.3-core/gl
go get -u github.com/go-gl/mathgl/mgl32
```
After getting all dependencies the project should work without any errors.

## Theory

This section describes the idea behind the different parts of this project. It is to note that most of the theory is taken from Eddie Lee's masterthesis. Sections that are taken from somewhere else will be mentioned explicitly.

### Infinite terrain

As mentioned earlier does the demo contain a seemingly infinite terrain. However nothing truely infinite could be computed by the PC. Instead the landscape is going to be divided into smaller chunks that themselves contain a squared grid of tiles. Each tile consists of two triangles. To contour the landscape a height-map is used that is repeated infinitely.

![infinite terrain](/assets/images/github/infinite_terrain.png)

Now only a small portion of the terrain is shown at once. A radius *r_i* around the camera is used to determine how many chunks are loaded around the camera. Every frame all chunks that are outside the radius *r_i* are being destroyed. Next missing chunks that are now inside the radius *r_i* are being created.

When creating the chunk, a grid of tiles is created. For each tile the heights *h1, h2, h3, h4* of the four vertices that make up the tile are being taken from the height-map. For each tile the position of the tile and the plane data of both triangles that make of the tile are being stored. The plane equation is ![plane equation](/assets/images/github/plane.png) with **p** *= (x y z)* being a point on the plane, **n** *= (A B C)* the normal of the plane. 
![plane distance](/assets/images/github/plane-dist.png) 
is the distance of the plane from the origin. The normal of a plane can be calculated by taking the cross product between the vertices of the tile. The two normals of both planes are 
![plane normal 1](/assets/images/github/plane-normal1.png) 
for the upper right plane and 
![plane normal 2](/assets/images/github/plane-normal2.png)  
for the lower right plane.

To speed up the check for chunks that have to be created, the current chunk *(px, pz)* the camera is in, is calculated. Only x- and z-coordinate are relevant. Then the radius in number of chunks is calculated as 
![chunk radius](/assets/images/github/chunk-radius.png) with *t_c* being the side length of a chunk. Then iterating from *(px-cx, pz-cz)* to *(px+cx, pz+cz)* and taking the current *x* and *z* position as a hash for a map that maps strings onto chunks. If the x-z-coordinate is not in the map it must mean that the respective chunk does not exist yet. If it is not existent the distance of this chunk to the chunk where the camera resides in is checked and if the distance is smaller than *r_i* then this chunk gets created and the coordinate of this newly created chunk is added together with the chunk to the map.

### View Frustum Culling

The used approach is the clip-space View Frustum Culling approach from [lighthouse3d.com](http://www.lighthouse3d.com/tutorials/view-frustum-culling/clip-space-approach-extracting-the-planes/). Using the model matrix **M** and the view matrix **V** of the camera, AABBs can be transformed into clip-space. Checking whether an AABB is inside the View Frustum is as easy as checking if the points of the transformed AABB are inside the clip-space which is now a cube. To check if a point is inside the View Frustum the point has to be inside all planes defined by the six sides of the View Frustum cube. If the point is on the outside of one plane then the point is outside of the View Frustum. 
To speed things up we only check two sides of the AABB. The point that is pointing the most in direction of the plane's normal and the point that is on the opposite side of the first point. If one point of the AABB is inside the View Frustum and the other one is outside then the AABB gets intersected by the View Frustum.

Using View Frustum Culling only chunks that are either inside the View Frustum or intersect it are collected. The plane data of all selected chunks are uploaded into a vertex array buffer and are actually used. By this chunks that cannot be seen by the camera are not considered and thus precious computing time can be used on chunks that actually might appear inside the current frame.

A point **p** gets converted into non-normaliced homogeneous coordinates by multiplying the point with **A** = **V** * **M**. **p**_t = **Ap**. Then the point **p**_t is transformed into normalized homogeneous coordinates by dividing all components of the point by its fourth component. ![clip-space homogeneous coordinates](/assets/images/github/clip-homo.png) is thus the point in normalized homogeneous clip-coordinates. For **p**_c to be inside the frustum the point has to be inside the canonical view volume meaning -1 < **p**_c,x < 1, -1 < **p**_c,y < 1 and -1 < **p**_c,z < 1 has to hold true. Then for a point in non-normalized homogeneous coordinates -**p**_t,w < **p**_t,x < **p**_t,w, -**p**_t,w < **p**_t,y < **p**_t,w and -**p**_t,w < **p**_t,z < **p**_t,w has to hold true. 

Now the requirements for all six planes of the homogeneous view frustum are being defined. Let the components of the matrix **A** be 

![clip-space matrix](/assets/images/github/clip-space-matrix.png).

As an example the restriction for the right plane is **p**_t,x < **p**_t,w. This can be written as -**p**_t,x + **p**_t,w > 0. Using the components of the matrix **A** the inequation becomes ![clip-space right plane](/assets/images/github/clip-right-plane.png). Extracting the components A,B,C,D of the plane yields the restriction (A B C D) = -**A**_1 + **A**_4 with **A**_1 and **A**_4 being column vectors of matrix **A**. This is done for all other five planes and thus the plane components for all six sides are

|plane  |restriction                    |
|-------|-------------------------------|
|left   |(A B C D) =   **A**_1 + **A**_4|
|right  |(A B C D) = - **A**_1 + **A**_4|
|bottom |(A B C D) =   **A**_2 + **A**_4|
|top    |(A B C D) = - **A**_2 + **A**_4|
|near   |(A B C D) =   **A**_3 + **A**_4|
|far    |(A B C D) = - **A**_3 + **A**_4|
 
The terrain is set up as a grid of axis aligned chunks. Each chunk has an AABB that is used to test against the View Frustum. Thus a collision test between the frustum and an AABB has to be performed for each chunk. The used approach only requires two points to be checked. Given a plane with a normal **n** the point that is furthest with regards to the direction of **n** is the point **p**_p. The other point **p**_n is on the opposite of **p**_p. Be the AABB defined by its position of the center of gravity and the two points **p**_min = (-w/2 -h/2 -d/2) and **p**_max = (w/2 h/2 d/2) with w,h and d being width, height and depth of the AABB. The algorithms for determining **p**_p and **p**_n are

```c
vec3 getPointP(vec3 pmin, vec3 pmax, vec3 n) {
    vec3 pp = pmin;
    if (n.x >= 0) pp.x = pmax.x;
    if (n.y >= 0) pp.y = pmax.y;
    if (n.z >= 0) pp.z = pmax.z;
    return pp;
}
```
```c
vec3 getPointN(vec3 pmin, vec3 pmax, vec3 n) {
    vec3 pn = pmax;
    if (n.x >= 0) pp.x = pmin.x;
    if (n.y >= 0) pp.y = pmin.y;
    if (n.z >= 0) pp.z = pmin.z;
    return pn;
}
```

For each plane the two points **p**_p and **p**_n are being determined and then used in the clip-space approach to see if they are inside. Is **p**_p outside of the current plane then the AABB is outside and the algorithm can return early. Is **p**_p inside and **p**_n is outside then the AABB intersects the View Frustum and the algorithm continues with the next plane. Did the AABB intersect non of the planes then the AABB is inside the View Frustum. The distance between the plane (**n** D) and a point **p** is defined as

```c
float planePointDistance(vec3 p, vec3 n, float D) {
    return n.Dot(p) + D;
}
```

The collision check between the View Frustum and a AABB is thus
```c
float checkFrustumAABB(Frustum frustum, AABB aabb) {
    int result = INSIDE;
    // clip space check for each plane
    for (int i = 0; i < 6; i++) {
        Plane plane = frustum.planes[i];
        vec3 n = plane.n;
        vec3 D = plane.D;
        vec3 pp = getPointP(aabb.min, aabb.max, n);
        vec3 pn = getPointN(aabb.min, aabb.max, n);
        if (planePointDistance(pp, n, D) < 0) return OUTSIDE;
        if (planePointDistance(pn, n, D) < 0) result = INTERSECT;
    }
    return result;
}
```

### Terrain rendering

The terrain collects all chunks that are inside the View Frustum or intersect it. Each chunk has two buffers. The first buffer contains the (x,z) center positions of all tiles in the chunk. The second buffer contains the plane data for each tile in the chunk that looks as follows
```c
struct PlaneData {
    vec4 plane1,
    vec4 plane2,
    float x,
    float z,
    vec2 padding
}
```
Here plane1 and plane2 are 4D-vectors (A B C D) with the plane data for both triangles that make up the tile.
![tile layout](/assets/images/github/tile.png)

The rendering of the terrain is straight forward. Each thread draws one tile. The vertex shader uses the tile positions and passes them through. In addition the id of the current vertex is passed to the geometry shader. The geometry shader creates two triangles with the position (x,z) and the width and depth of the tile defined by the terrain. To get the height of four points v_1, v_2, v_3 and v_2, v_3, v_2, v_4 the plane data of the respective triangle is used by solving the equation ![plane height](/assets/images/github/plane-height.png). 

The fragment shader then just uses flat shading to color the whole terrain black.

### Wind Simulation

To give the user more interaction the movement of the camera induces a force on the grass field. The direction of the force is provided by the direction the camera is moving. The wind simulation uses two vector fields that are centered around camera. One vector field specifies the wind acceleration which is strong close to the camera and decreases the further away the cell is from the camera. The other vector field contains the wind velocities. In this case both vector fields are of the same size and have *2N+1* cells in both directions. The velocity vector field is initialized with zero vectors while the acceleration vector field stays constant and is defined as follows.

![acceleration vector field plot](/assets/images/github/acceleration-vector-field.png)
The acceleration vector field ranges from *-N* to *N* in both directions with cell *(0,0)* being the center of the vector field. The vectors of the vector field should point away from the center of the vector field. This can be achieved by *f(x,y) = (x y)* however does this lead to vectors that are further away from the center are having bigger magnitudes. The idea is that acceleration vectors are bigger the closer they are to the center and decrease towards the edges of the vector field as can be seen in the figure above. The used formula to calculate the acceleration vector field is ![acceleration vector field formula](/assets/images/github/acceleration-formula.png). The *x* and *y* values are calculated by ![acceleration vector field indices](/assets/images/github/acceleration-indices.png) where *x_g* and *y_g* are the indices of the vector field with *(0, 0)* being the center of the vector field and *s* being a spread value that determines the range of *x* and *y* values in the calculation for the acceleration vectors. Using a bigger value of *s* results in smaller magnitudes at the borders of the vector field until those vectors are zero eventually.

The velocity vector field is updated each frame. The velocity vector field is also always centered around the position of the camera. However if the camera moves from one tile to the next the velocities from the previous frame must be shifted to yield believable behaviour of the wind. Thus the center of the previous frame *(x_p, y_p)* is saved and is subtracted from the position of the current center *(x_c, y_c)*. Thus the velocity offset is *(dx, dy) = (x_c, y_c) - (x_p, y_p)*. The velocity force field calculation takes places on the GPU. The velocity vector field is always centered around the camera position of the current frame. For each thread that calculates the velocity vector at position *(x_t, y_t)* grabs the velocity of the previous frame *v_p* at the position *(x_n, y_n) = (x_t, y_t) + (dx, dy)*. If this position *(x_n, y_n)* is outside of the vector field, then ***v**_p* is the zero vector as the wind outside of the velocity vector field is considered non existent. 

![acceleration vector field formula](/assets/images/github/acceleration-vector-field-view-dependent.png)

Also relevant for the calculation of the new velocity is the acceleration *a* at *(x_c, y_c)*. However we only take acceleration vectors in account that are similiar to the movement direction of the camera. Let ***d**_c* be the normalized 2D direction of the camera from the last frame to the current frame. We will use the direction dependent acceleration ![acceleration view dependent](/assets/images/github/acceleration-view-dependent.png). Values of ***a**_d* are clamped between 0 and 1 to ignore vectors that face in the opposite direction of the view vector. The resulting vector field can be seen above. Here ***d**_c* is shown in blue while the red vectors are the view dependent acceleration vectors.

The new velocity is calculated by ***v**_n = d_w **v**_p + **a**_d* with *d_w* being a damping constant that gradually slows down the wind velocity. 

### Grass rendering and simulation

Each tile of the terrain is going to have multiple grass blades. As this will result in a big workload some optimizations have to be considered. During setup one buffer with random (x z) root positions of the grass blades is generated. This buffer contains all positions for one tile. It will be reused for each tile. This means that instanced rendering is going to be used. The number of instances equals the number of tiles that are actually rendered. Those tiles had been determined during the update of the terrain. Besides the root positions buffer, the wind velocities as well as the terrain plane data buffer are used. 

The vertex shader simply passes the root positions through as well as the instance ID that is used to identify which plane data to use.

#### Creating the grass blade

The geometry shader is responsible for creating the grass blades based on the root positions. First using the plane data for the current tile the height of the current grass root position can be determined. A tile consists of two triangles with different slopes. Is x < z then the root position is on the upper left triangle else its on the lower right triangle. Similar as for the terrain, the y component of the grass root position can be calculated with ![plane height](/assets/images/github/plane-height.png). Next the LOD is calculated for the current tile. The distance of the center position of the tile to the camera is used to calculate the LOD with ![lod](/assets/images/github/lod.png), where *d_n* is the normalized distance of the camera to the tile center and *γ* determines how fast the function converges to zero. The LOD value will reach from 0 to 3 where close tiles will be of LOD level 3 and far away tiles LOD level 0.

![lod](/assets/images/github/segments.png)

Based on the LOD, the grass blades are going to have more or less detailed geometry. LOD level 3 has grass blades with 5 segments, LOD level 2 has 3 segments, LOD level 1 has 1 segment and LOD level 0 only creates a rectangle of two triangles for the whole tile.

#### Wind

Besides the detail of the grass blades *N_blade* also the number of grass blades per tile based on the LOD level is decreased.

|LOD| blade count      |
|---|------------------|
| 3 | *N_blade*        |
| 2 | *0.95 * N_blade* |
| 1 | *0.85 * N_blade* |
| 0 | *1*              |

The grass should be influenced by the wind. The wind velocity force field is always centered at the position of the camera and the cells of the force field align with the tiles of the terrain. Using the wind velocity **v** for the current tile all segments of the grass blade are going to be displaced by the velocity of the wind. Each vertex of position **p** is displaced by

![lod](/assets/images/github/displaced-pos.png),

where *c = 1 - v* and *v* in this case is the texture coordinate of the current vertex. In addition to displacing the grass blades based on the wind velocity the blades should wave in the wind. There is a proportional relationship between the bending of the grass blade and the waving area. Assuming that the wind is only blowing in the x-z space, bended grass blades have a smaller area that is affected by the wind than grass blades that are straight pointing upwards. This means that area of occilation decreases the stronger the grass blade is bended.

![lod](/assets/images/github/grass-bend.png)

The first part of the equation decribes the difference between the root position height of the grass blade **r** and the height of the current vertex **p**. The second part takes the magnitude of the wind velocity and scales it by a bending constant *c_b* between 0 and 1. The time function is a sinusoidal function that models the waving of the grass depending the current time *t*. By using the displaced position and then applying the following function the final position of the grass blade vertex is calculated.

![lod](/assets/images/github/time-bend.png)

#### Randomization

At this point each grass blade of one blade looks the same which creates a very artificial look. Thus randomization is introduced at several places to make the grass look more natural. Usually a noise texture could be used that each grass blade samples from but each grass root position ***r*** is already randomized in its *x* and *z* position, so those positions could be used. In addition is the position of the tile ***t*** used as well so that grass blades in neighboring tiles look distinct from each other. A random value *r* can be calculated by ![lod](/assets/images/github/rand.png). Here the function f returns the fractional part of a number.

Now using this random number several parts of the grass generation can be randomized. First the root positions of the grass are the same in each tile. To change this each grass root position can be shifted in *x* and *z* direction. However both positions must be in the range from 0 to 1. Thus the *x* and *z* components are looped so that they are always in this range.

Next the appearance of the grass blade is going to be randomized. This will affect the width, height and the used texture of the grass blade. The latter will be encoded with a texture ID. Three different textures had been used for the grass in different stages of its life. A fourth texture is used for the LOD level 0 tile.

Lastly the waving of the grass blade is randomized by assigning random frequencies and phase shifts ![lod](/assets/images/github/rand-time.png). The *%* is the modulo operator that returns the remainder of the division *t / f*.

#### Grass rendering

The rendering of the grass uses two textures. The first texture is determined by the texture ID assigned randomly in the geometry shader. Here one of three color textures is chosen. The second texture is a alpha-texture that determines if a fragment is discarded. This technique is called alpha-to-coverage and it basically cuts out the part of color texture that contains the grass. However is the second texture only used for LOD levels 1 to 3 as LOD level 0 is the tile with a special color texture.

After sampling from the color texture grass gets specular lighting. To emulate ambient occlusion that makes grass blades darker at the bottom the resulting color of the shading is linearily interpolated with black using the *v* component of the uv coordinate of the current fragment ![lod](/assets/images/github/grass-ao.png). Thus the final color is ![lod](/assets/images/github/grass-shaded.png).

### Postprocessing

#### Fog

Having an infinite View Frustum is impossible also having a big loading radius for the terrain is costly. To get away with a smaller loading radius fog can be used to let the terrain merge with the color of the background. In addition is the fog used to emulate the effect that far away objects have a blueish tint. 

The effect occures because of athmospheric scattering. The fog is a medium that scatters light. Some of the bluish light of the sky is scattered towards the camera. This is called *in-scattering*. The greater the distance to an object the more *in-scattering* occurs. It doesn't have to be bluish. On an overcast day the athmospheric scattering can be grayish. In addition scattering near the sun can be a yellowish tint.

The algorithm for calculating the fog had been taken from [www.iquilezles.org](http://www.iquilezles.org/www/articles/fog/fog.htm). Two intensities will be calculated. The first one describes the influence of the fog. The greater the distance *d_f* between the object and the camera the more fog is in between and the more athmospheric scattering occures. This intensity is modelled by ![fog](/assets/images/github/fog-intensity.png), with *b* being the density of the fog. The other intensity is the influence of the sun color. The closer the camera is looking towards the sun the more the fog turns into a yellowish tint. This can be modelled by taking the dot product between the normalized view direction of the camera *d_cam* and the normalized direction of the camera to the sun *d_sun*, thus the sun intensity is ![sun intensity](/assets/images/github/sun-intensity.png). By using the maximum function, cases in where the camera looks in the opposite direction to the sun are ignored. 

To calculate the color of the fog depending to the view direction of the camera with respect to the sun can be calculated as linear interpolation between the color of the sky ***c**_sky* and the color of the sun ***c**_sun*. The color of the fog is then ![fog color](/assets/images/github/fog-color.png) where σ describes the spread of the sun color where higher values lead to a smaller influence of the sun color. 

Finally the color of the fog is interpolated with the color of the object ***c**_obj* depending on the fog intensity. The resulting color ***c**_res* is ![fog color](/assets/images/github/result-color.png).

#### Depth of Field (DoF)

A real camera can only show objects of a certain distance in acceptable sharpness. This depends on the aperture size, the focal length of the camera lense and the general distance to the object in question. To model the camera properties a distance where the image is completely sharp and a range where the objects are still acceptably sharp.

First the image *I* is blurred using gaussian blur *I_gauss*. Taking the approach from [http://encelo.netsons.org](http://encelo.netsons.org/2008/04/15/depth-of-field-reloaded/) blur intensity can be calculated by using the distance to the fragment *d_f*, the focal range of the lense *f* and the range of acceptable sharpness *s* yielding ![blur intensity](/assets/images/github/blur.png). Using the blur intensity the depth of field color ***c**_res* can be calculated by ![blur color](/assets/images/github/dof-color.png). Here ***c**_b* is the color of the blurred image and ***c**_f* the color of the original image.

#### Bloom

Bloom is also an camera artifact where areas of bright light bleed over dark edges. It is because the camera lens convolves the light with an airy disk.

The bloom is calculated in two steps. The first step takes the initial image and calculates the luminocity and thresholding pixels to ignore pixels that are below this threshold *t*. The luminocity *l* of a pixel of color ***c**_f* is calculated by 

![blur color](/assets/images/github/luminocity.png)

where the second vector approximates the eye's sensibility towards different wave spectra. The eye has three different cone types to see colors. Those cones are sensitive to different kinds of wavelengths. The red and green receptors are overlapping each other in the green to red wavelengths which makes the eye more sensitive to green and red than to blue which is mostly covered by the blue receptor.

The following image is taken from [wikipedia](https://en.wikipedia.org/wiki/Color_vision#/media/File:Cone-fundamentals-with-srgb-spectrum.svg).
![normalized cone spectra](/assets/images/github/cone-spectra.png)

Then the luminocity is then thresholded with the function 

![normalized cone spectra](/assets/images/github/luminocity-threshold.png)

to create a greater effect between bright and dark regions. Then this luminocity image that is the result of the first step gets blurred using gaussian blur to create the effect of light spilling over dark borders. In the second step the color of the original image ***c**_f* and the luminocity are simply added together 

![normalized cone spectra](/assets/images/github/bloom-color.png).
