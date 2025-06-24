document.addEventListener("DOMContentLoaded", () => {
    const fadeElements = document.querySelectorAll("body *");

    const fadeInOnScroll = () => {
        fadeElements.forEach(element => {
            const rect = element.getBoundingClientRect();
            if (rect.top < window.innerHeight * 1.5) {
                element.style.opacity = 1;
                element.style.transform = "translateY(0)";
            }
        });
    };

    window.addEventListener("scroll", fadeInOnScroll);
    fadeInOnScroll(); 
});


function contributionValidate() {
    const form = document.forms["myform"];
    const skillName = form["tb1"].value.trim();
    const skillDesc = form["tb2"].value.trim();
    const fileInput = form["tb6"];

    const skillNameRegex = /^[a-zA-Z0-9 ]{5,}$/; 
    const allowedFileTypes = /(\.mp4|\.mp3|\.wav|\.avi|\.mov|\.pdf|\.docx|\.txt)$/i; 

    if (!skillNameRegex.test(skillName)) {
        alert("Skill Name must be at least 5 characters long and contain only letters, numbers, and spaces.");
        return false;
    }

    if (skillDesc.length < 25) {
        alert("Skill Description should be greater than 25 characters.");
        return false;
    }

    const file = fileInput.files[0];
    if (!allowedFileTypes.test(file.name)) {
        alert("Invalid file type. Allowed types are: mp4, mp3, wav, avi, mov, pdf, docx, txt.");
        return false;
    }

    return true;
}


    function revealInfo(){
        document.getElementById('skillform-div').style.display="none";
        document.getElementById('contribution-info').style.display="block";
    }
    function hideInfo(){
        document.getElementById('skillform-div').style.display="block";
        document.getElementById('contribution-info').style.display="none";
    }

    let signup=document.getElementById('signupform');
    let signin=document.getElementById('signinform');
    function hideI(){
        signup.style.display='none';
        signin.style.display='flex';
    }
    function hideII(){
        signup.style.display='flex';
        signin.style.display='none';
    }
    function validateI(){
        let emailpatt = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        let pwdpatt = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,12}$/;
        let email = document.forms["myformI"]["tb1"].value;
        let pwd = document.forms["myformI"]["tb2"].value;
        if(!emailpatt.test(email)){
            alert('Email does not match the standard criteria!');
            return false;
        }
        if(!pwdpatt.test(pwd)){
            alert('Password does not match the standard criteria!');
            return false;
        }
        return true; 
    }
    function validateII(){
        let emailpatt = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        let pwdpatt = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,12}$/;
        let fullnamepatt = /^[A-Z][a-z]+(?: [A-Z][a-z]+)*$/;
        let addrpatt = /^[A-Za-z][A-Za-z0-9,. ]{8,98}[A-Za-z0-9]$/;
        
        let fullname = document.forms["myformII"]["tb1"].value.trim();
        let email = document.forms["myformII"]["tb2"].value.trim();
        let addr = document.forms["myformII"]["tb4"].value.trim();
        let contact = document.forms["myformII"]["tb5"].value.trim();
        let pwd = document.forms["myformII"]["tb6"].value.trim();

        if(!fullnamepatt.test(fullname) || fullname.length < 10 || fullname.length >= 25){
            alert('Fullname does not match the standard criteria!');
            return false;
        }
        if(!emailpatt.test(email)){
            alert('Email does not match the standard criteria!');
            return false;
        }
        if(!addrpatt.test(addr) || addr.length < 10){
            alert('Address does not match the standard criteria!');
            return false;
        }
        if(contact.length < 10 || contact.length>10){
            alert('Contact No does not match the standard criteria!');
            return false;
        }
        if(!pwdpatt.test(pwd)){
            alert('Password does not match the standard criteria!');
            return false;
        }
        return true;
    }
    function OTPvalidate()
    {
        let otp=document.forms["myform"]["tb"].value.trim();
        let otppatt=/^\d{6}$/;
        if(!otppatt.test(otp)){
            alert('OTP does not include numbers or OTP length exceeds limit 6!');
            return false;
        }
        return true;
    }
    function UPIFunc(){
        document.getElementById('UPIDiv').style.display='block';
        document.getElementById('QRDiv').style.display='none';
        document.getElementById('debitcreditDiv').style.display='none';
    }
    function QRFunc(){
        document.getElementById('UPIDiv').style.display='none';
        document.getElementById('QRDiv').style.display='block';
        document.getElementById('debitcreditDiv').style.display='none';
    } 
    function DCFunc(){
        document.getElementById('UPIDiv').style.display='none';
        document.getElementById('QRDiv').style.display='none';
        document.getElementById('debitcreditDiv').style.display='block';
    }
    function myfunc(){
        //var sk=document.getElementByID('skillcoins').value();
        document.getElementById('skprice').innerHTML='₹50';
        console.log(document.getElementByID('skillcoins').value());
    }

    function myhideI(){
        document.getElementById('info').style.color='#e05e00';
        document.getElementById('skills').style.color='#000';
        document.getElementById('wallet').style.color='#000';
        document.getElementById('guide').style.color='#000';
        document.getElementById('div1').style.display='block';
        document.getElementById('div2').style.display='none';
        document.getElementById('div3').style.display='none';
        document.getElementById('div4').style.display='none';
    }
    function myhideII(){
        document.getElementById('info').style.color='#000';
        document.getElementById('skills').style.color='#e05e00';
        document.getElementById('wallet').style.color='#000';
        document.getElementById('guide').style.color='#000';
        document.getElementById('div1').style.display='none';
        document.getElementById('div2').style.display='block';
        document.getElementById('div3').style.display='none';
        document.getElementById('div4').style.display='none';
    }
    function myhideIII(){
        document.getElementById('info').style.color='#000';
        document.getElementById('skills').style.color='#000';
        document.getElementById('guide').style.color='#000';
        document.getElementById('wallet').style.color='#e05e00';
        document.getElementById('div1').style.display='none';
        document.getElementById('div2').style.display='none';
        document.getElementById('div3').style.display='block';
        document.getElementById('div4').style.display='none';
    }
    function myhideIV(){
        document.getElementById('info').style.color='#000';
        document.getElementById('skills').style.color='#000';
        document.getElementById('guide').style.color='#e05e00';
        document.getElementById('wallet').style.color='#000';
        document.getElementById('div1').style.display='none';
        document.getElementById('div2').style.display='none';
        document.getElementById('div3').style.display='none';
        document.getElementById('div4').style.display='block';
        
    }
    function showUpdateForm(){
        document.getElementById('editformopenbtn').style.display='none';
        document.getElementById('editformclosebtn').style.display='block';
        document.getElementById('editProfileForm').style.display='block';
    }
    function closeUpdateForm(){
        document.getElementById('editformopenbtn').style.display='block';
        document.getElementById('editformclosebtn').style.display='none'; 
        document.getElementById('editProfileForm').style.display='none';
    }
    function profileValidation() {
const name = document.forms['myform']['tb1'].value;
const nameRegex = /^[A-Z][a-zA-Z\s]*$/;
if (!nameRegex.test(name) || !/^[A-Z]/.test(name.trim().split(' ')[0])) {
    alert("Please enter a valid name (only alphabets, spaces, and first letter of each word capitalized).");
    return false;
}

const address = document.forms['myform']['tb2'].value;
const addressRegex = /^[a-zA-Z\s.,]{15,}$/;
if (!addressRegex.test(address)) {
    alert("Please enter a valid address (greater than 15 characters, and only letters, spaces, commas, and periods allowed).");
    return false;
}

const contactNo = document.forms['myform']['tb3'].value;
const contactNoRegex = /^\d{10}$/;
if (!contactNoRegex.test(contactNo)) {
    alert("Please enter a valid contact number (exactly 10 digits).");
    return false;
}

const bio = document.forms['myform']['tb4'].value;
const bioRegex = /^[a-zA-Z\s.,!]+$/;
if (!bioRegex.test(bio)) {
    alert("Please enter a valid bio (only alphabets, spaces, and special characters like . , ! allowed).");
    return false;
}

const profilePicture = document.forms['myform']['tb5'].value;
const imageExtensions = /\.(jpg|jpeg|png|bmp)$/i;
if (!imageExtensions.test(profilePicture)) {
    alert("Please upload a valid image file (jpg, jpeg, png, bmp).");
    return false;
}

return true;
}

function changeStateImg(){
    let state = document.getElementById("states").value;
    let stateImg = document.getElementById("stateImg");
    let stateDesc = document.getElementById("stateDesc");
    
    switch(state){
        case "andhra_pradesh": stateImg.src = "/static/img/states/andhra.avif"; stateDesc.innerHTML = "Andhra Pradesh is known for its rich cultural heritage, historic temples, and scenic landscapes. Visit the famous Tirupati Temple, explore the ancient ruins of Lepakshi, or relax at the picturesque Araku Valley."; break;
        case "arunachal_pradesh": stateImg.src = "/static/img/states/arunachalpradesh.avif"; stateDesc.innerHTML = "Arunachal Pradesh is a paradise for nature lovers with its serene monasteries, lush valleys, and unexplored forests. Tawang Monastery and Ziro Valley are must-visit attractions."; break;
        case "assam": stateImg.src = "/static/img/states/assam.avif"; stateDesc.innerHTML = "Assam is famous for its tea gardens, wildlife, and spiritual sites. Explore the Kaziranga National Park, visit the Kamakhya Temple, or enjoy a river cruise on the Brahmaputra."; break;
        case "bihar": stateImg.src = "/static/img/states/bihar.avif"; stateDesc.innerHTML = "Bihar boasts a rich historical past with sites like Nalanda University, Bodh Gaya, and Rajgir. It is a major Buddhist pilgrimage center with deep spiritual significance."; break;
        case "chhattisgarh": stateImg.src = "/static/img/states/chhattisgarh.jpg"; stateDesc.innerHTML = "Chhattisgarh is known for its tribal culture, waterfalls, and ancient temples. Visit the Chitrakote Falls, explore the caves of Bastar, or experience the traditional tribal dance."; break;
        case "goa": stateImg.src = "/static/img/states/goa.avif"; stateDesc.innerHTML = "Goa is India's party capital with beautiful beaches, Portuguese heritage, and vibrant nightlife. Explore the churches of Old Goa, enjoy watersports, or relax at Palolem Beach."; break;
        case "gujarat": stateImg.src = "/static/img/states/gujrat.png"; stateDesc.innerHTML = "Gujarat is known for its vibrant festivals, heritage sites, and wildlife. Visit the Gir National Park, explore the Rann of Kutch, or experience the colorful Navratri celebrations."; break;
        case "haryana": stateImg.src = "/static/img/states/haryana.avif"; stateDesc.innerHTML = "Haryana is famous for its historical sites, lakes, and traditional crafts. Visit the Sultanpur Bird Sanctuary, explore the Kurukshetra battle site, or enjoy rural tourism in the villages."; break;
        case "himachal_pradesh": stateImg.src = "/static/img/states/himachal.avif"; stateDesc.innerHTML = "Himachal Pradesh is a haven for adventure lovers and nature enthusiasts. Explore Shimla, trek to Spiti Valley, or visit the famous temples in Manali."; break;
        case "jharkhand": stateImg.src = "/static/img/states/jharkhand.jpg"; stateDesc.innerHTML = "Jharkhand is home to rich tribal culture and natural beauty. Visit Betla National Park, explore Hundru Falls, and experience the vibrant tribal festivals."; break;
        case "karnataka": stateImg.src = "/static/img/states/karnataka.jpg"; stateDesc.innerHTML = "Karnataka offers a mix of heritage and modernity. Visit Mysore Palace, explore Hampi's ruins, or enjoy the coffee plantations of Coorg."; break;
        case "kerala": stateImg.src = "/static/img/states/kerala.avif"; stateDesc.innerHTML = "Kerala is known as 'God’s Own Country' with its serene backwaters, lush greenery, and Ayurvedic retreats. Explore Munnar, visit Alleppey, or enjoy a Kathakali performance."; break;
        case "madhya_pradesh": stateImg.src = "/static/img/states/madhyapradesh.avif"; stateDesc.innerHTML = "Madhya Pradesh: Known as the 'Heart of India,' Madhya Pradesh is home to historical wonders like Khajuraho temples, Sanchi Stupa, and the wildlife-rich Kanha National Park."; break;
        case "maharashtra": stateImg.src = "/static/img/states/Maharashtra.webp"; stateDesc.innerHTML = "Maharashtra: Famous for Mumbai, Maharashtra boasts heritage sites like Ajanta & Ellora Caves, forts of Shivaji, and the serene beaches of Konkan."; break;
        case "manipur": stateImg.src = "/static/img/states/manipur.jpg"; stateDesc.innerHTML = "Manipur: A jewel of Northeast India, Manipur is known for its floating Keibul Lamjao National Park, Loktak Lake, and classical Manipuri dance."; break;
        case "meghalaya": stateImg.src = "/static/img/states/meghalaya.avif"; stateDesc.innerHTML = "Meghalaya: The 'Abode of Clouds' offers living root bridges, mesmerizing waterfalls, and caves, with Cherrapunji and Shillong being prime attractions."; break;
        case "mizoram": stateImg.src = "/static/img/states/mizoram.avif"; stateDesc.innerHTML = "Mizoram: A land of rolling hills and bamboo forests, Mizoram showcases unique tribal culture, with Reiek Tlang and Vantawng Falls being major draws."; break;
        case "nagaland": stateImg.src = "/static/img/states/nagaland.jpg"; stateDesc.innerHTML = "Nagaland: Home to the vibrant Hornbill Festival, Nagaland is known for its tribal heritage, scenic hills, and World War II historical sites."; break;
        case "odisha": stateImg.src = "/static/img/states/odisha.png"; stateDesc.innerHTML = "Odisha: The state is famous for the Sun Temple of Konark, the Jagannath Puri temple, and the vast Chilika Lake with its migratory birds."; break;
        case "punjab": stateImg.src = "/static/img/states/punjab.avif"; stateDesc.innerHTML = "Punjab: Rich in Sikh heritage, Punjab is home to the Golden Temple, vibrant festivals, and the historical Wagah Border ceremony."; break;
        case "rajasthan": stateImg.src = "/static/img/states/rajasthan.webp"; stateDesc.innerHTML = "Rajasthan: A land of palaces and deserts, Rajasthan is known for Jaipur's forts, Jaisalmer's sand dunes, and Udaipur's lakes."; break;
        case "sikkim": stateImg.src = "/static/img/states/sikkim.avif"; stateDesc.innerHTML = "Sikkim: A picturesque state with breathtaking monasteries, Sikkim offers Kanchenjunga views, adventure sports, and vibrant Buddhist culture."; break;
        case "tamil_nadu": stateImg.src = "/static/img/states/tamilnadu.avif"; stateDesc.innerHTML = "Tamil Nadu: Home to majestic temples, Tamil Nadu boasts the shore temples of Mahabalipuram, Meenakshi Temple, and the Nilgiri Hills."; break;
        case "telangana": stateImg.src = "/static/img/states/telangana.avif"; stateDesc.innerHTML = "Telangana: Known for the Charminar, Golconda Fort, and Ramoji Film City, Telangana showcases a blend of heritage and modernity."; break;
        case "tripura": stateImg.src = "/static/img/states/tripura.jpg"; stateDesc.innerHTML = "Tripura: This small state is famous for the Ujjayanta Palace, Neermahal water palace, and its lush green landscapes."; break;
        case "uttar_pradesh": stateImg.src = "/static/img/states/uttarpradesh.avif"; stateDesc.innerHTML = "Uttar Pradesh: A land of spiritual significance, Uttar Pradesh is home to the Taj Mahal, Varanasi’s ghats, and Mathura’s Krishna temples."; break;
        case "uttarakhand": stateImg.src = "/static/img/states/uttarakhand.avif"; stateDesc.innerHTML = "Uttarakhand: The 'Dev Bhoomi' is known for the Char Dham Yatra, the scenic Nainital, and adventure trekking in the Himalayas."; break;
        case "west_bengal": stateImg.src = "/static/img/states/westbengal.webp"; stateDesc.innerHTML = "West Bengal: A state rich in history, art, and culture, West Bengal is famous for Kolkata’s Victoria Memorial, Darjeeling’s tea gardens, and the Sundarbans."; break;
        default: stateImg.src = ""; stateDesc.innerHTML = "";
    }
}


function validateIII() {
    let form = document.forms['myformII'];
    
    let experience = form['tb1'].value;
    if (!/^[0-9]+$/.test(experience) || parseInt(experience) < 1) {
        alert('Experience must be a valid number and at least 1 year.');
        return false;
    }
    
    let languages = form['tb2'].value;
    if (!/^[A-Za-z, ]+$/.test(languages)) {
        alert('Languages should contain only alphabets and commas.');
        return false;
    }
    
    let charges = form['tb4'].value;
    if (!/^[0-9]+$/.test(charges) || parseInt(charges) <= 0) {
        alert('Expected charges must be a valid positive number.');
        return false;
    }
    
    let idProof = form['tb5'].files[0];
    if (idProof) {
        let idProofExt = idProof.name.split('.').pop().toLowerCase();
        if (!['pdf', 'png', 'jpg', 'jpeg'].includes(idProofExt)) {
            alert('Identity proof must be a PDF or image file (PNG, JPG, JPEG).');
            return false;
        }
    } else {
        alert('Please upload an identity proof.');
        return false;
    }
    
    let bio = form['tb6'].value;
    if (bio.length < 20) {
        alert('Short Bio must be at least 20 characters long.');
        return false;
    }
    
    return true;
}

function bookingRejected() {
    let reason = prompt("Enter the reason for rejecting the request:");
  
    if (reason && reason.trim() !== "") {
      document.getElementById('msg').value = reason;
      document.getElementById('rejectForm').submit();
    } else {
      alert("Reason cannot be empty.");
    }
  }
  
 function bookingAccepted(){
    document.getElementById('acceptForm').submit();
 }

function forumValidate(){
    const msgPatt = /^[a-zA-Z0-9& ,.!@?;]{15,100}$/;
    let msg = document.forms["myforumform"]["tb1"].value.trim();
    if(!msgPatt.test(msg)){
        alert('Message is too short, should be atleast 15 characters long!');
        return false;
    }
    return true;
}
function redirecttoSell(){
    const form = document.createElement("form");
    form.method="POST";
    form.action="/sellartpieces";
    document.body.appendChild(form);
    form.submit();
}
function validateSellForm() {
    const namePatt = /^([A-Z][a-z]+)(\s?[A-Z][a-z]+)*$/;
    const descPatt = /^[a-zA-Z\s.,-]{15,100}$/;
    const pricePatt = /^[0-9]+(\.[0-9]{1,3})?$/;
    const validExtensions = ['.jpg', '.jpeg', '.png', '.webp', '.svg'];
    const maxSizeInBytes = 1 * 1024 * 1024;

    const name = document.getElementsByName("tb1")[0].value.trim();
    const desc = document.getElementsByName("tb2")[0].value.trim();
    const price = document.getElementsByName("tb3")[0].value.trim();
    const imageInput = document.getElementsByName("tb4")[0];
    const file = imageInput.files[0];
    const fileName = imageInput.value.toLowerCase();

    if (!namePatt.test(name) || name.length < 5 || name.length > 30) {
        alert('Product name should start with a capital letter and be 5–30 characters long, containing only alphabets and optional spaces.');
        return false;
    }

    if (!descPatt.test(desc)) {
        alert('Product description should be 15–100 characters long and contain only letters, spaces, commas, periods, or hyphens.');
        return false;
    }

    if (!pricePatt.test(price)) {
        alert('Price should be a valid number (e.g., 100 or 99.99).');
        return false;
    }

    if (!file) {
        alert("Please select an image file.");
        return false;
    }

    const isValidExtension = validExtensions.some(ext => fileName.endsWith(ext));
    if (!isValidExtension) {
        alert("Image must be in one of the following formats: " + validExtensions.join(', '));
        return false;
    }

    if (file.size > maxSizeInBytes) {
        alert("Image size must not exceed 1 MB.");
        return false;
    }

    return true;
}

