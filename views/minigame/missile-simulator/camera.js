(function(root){
  class Camera {
    constructor(){this.scale=.22;this.x=0;this.width=1000;this.height=480;this.worldWidth=6800;}
    get y(){return 1180-(this.height-80)/this.scale;}
    get maxX(){return Math.max(0,this.worldWidth-this.width/this.scale);}
    resize(width,height){this.width=width;this.height=height;this.pan(this.x);}
    pan(x){this.x=Math.max(0,Math.min(this.maxX,x));}
    zoom(scale){const center=this.x+this.width/(2*this.scale);this.scale=Math.max(.04,Math.min(.65,scale));this.pan(center-this.width/(2*this.scale));}
    fit(){this.zoom(this.width/this.worldWidth);this.pan(0);}
    screen(x,y){return {x:(x-this.x)*this.scale,y:(y-this.y)*this.scale};}
    world(x,y){return {x:x/this.scale+this.x,y:y/this.scale+this.y};}
  }
  root.NavalCamera=Camera;
  if(typeof module!=='undefined')module.exports=Camera;
})(typeof window!=='undefined'?window:globalThis);
